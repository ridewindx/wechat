package mch

import (
	"io"
	"bytes"
	"unicode"
	"os"
)

type BillType string

const (
	BillTypeALL             BillType = "ALL"             // 返回当日所有订单信息，默认值
	BillTypeSUCCESS         BillType = "SUCCESS"         // 返回当日成功支付的订单
	BillTypeREFUND          BillType = "REFUND"          // 返回当日退款订单
	BillTypeRECHARGE_REFUND BillType = "RECHARGE_REFUND" // 返回当日充值退款订单（相比其他对账单多一栏“返还手续费”）
)

type DownloadBillRequest struct {
	// 必选参数
	BillDate string // 下载对账单的日期，格式：20140603
	BillType BillType // 账单类型

	// 可选参数
	DeviceInfo string // 微信支付分配的终端设备号，填写此字段，只下载该设备号的对账单
	TarType    bool // 是否压缩账单
}

func (client *Client) DownloadBill(req *DownloadBillRequest, filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		if err != nil {
			os.Remove(filename)
		}
	}()
	return client.DownloadBillToWriter(req, file)
}

func (client *Client) DownloadBillToWriter(req *DownloadBillRequest, writer io.Writer) error {
	reqMap := make(map[string]string, 8)
	reqMap["bill_date"] = req.BillDate
	if req.BillType != "" {
		reqMap["bill_type"] = string(req.BillType)
	}
	if req.DeviceInfo != "" {
		reqMap["device_info"] = req.DeviceInfo
	}
	if req.TarType {
		reqMap["tar_type"] = "GZIP"
	}

	buffer, err := client.makeRequest(reqMap, false)
	defer pool.Put(buffer)
	if err != nil {
		return err
	}

	url := API_URL + "/pay/downloadbill"
	_, err = client.postToWriter(url, buffer, writer)
	if err != nil {
		bizError, ok := err.(*BizError)
		if !ok || bizError.ErrCode != ErrCodeSYSTEMERROR  {
			return err
		}
		url = switchReqURL(url)
		_, err = client.postToWriter(url, buffer, writer) // retry
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) postToWriter(url string, body io.Reader, writer io.Writer) (written int64, err error) {
	repBody, err := client.post(false, url, body)
	if err != nil {
		return 0, err
	}
	defer repBody.Close()

	buffer := make([]byte, 32<<10) // 与 io.copyBuffer 里的默认大小一致
	switch n, err := io.ReadFull(repBody, buffer); err {
	case nil:
		written, err := bytes.NewReader(buffer).WriteTo(writer)
		if err != nil {
			return written, err
		}
		n, err := io.CopyBuffer(writer, repBody, buffer)
		written += n
		return written, err
	case io.ErrUnexpectedEOF:
		content := buffer[:n]
		bs := bytes.TrimLeftFunc(content, func(r rune) bool {
			return unicode.IsSpace(r)
		})
		if bytes.HasPrefix(bs, []byte("<xml>")) {
			_, err = client.toMap(bytes.NewReader(bs), false)
			if err != nil {
				return 0, err
			}
		}
		return bytes.NewReader(content).WriteTo(writer)
	case io.EOF:
		return 0, nil
	default:
		return 0, err
	}
}
