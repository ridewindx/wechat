package mp

import "strconv"

type CorpDepartment struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ParentID int64  `json:"parentid"`
	Order    int64  `json:"order"`
}

func (c *Client) GetCorpDepartmentList(id ...int64) ([]CorpDepartment, error) {
	u := CORP_BASE_URL.Join("/department/list")
	if len(id) > 0 {
		u = u.Query("id", strconv.FormatInt(id[0], 10))
	}

	var result struct{
		Err
		Departments []CorpDepartment `json:"department"`
	}
	err := c.Get(u, &result)
	if err != nil {
		return nil, err
	}
	return result.Departments, nil
}

type CorpUser struct {
	ID         string  `json:"userid"`
	Name       string  `json:"name"`
	Department []int64 `json:"department"`
}

func (c *Client) GetCorpUserList(departmentID int64, fetchChild ...bool) ([]CorpUser, error) {
	u := CORP_BASE_URL.Join("/user/simplelist").Query("department_id", strconv.FormatInt(departmentID, 10))
	if len(fetchChild) > 0 && fetchChild[0] {
		u = u.Query("fetch_child", "1")
	}

	var result struct{
		Err
		Users []CorpUser `json:"userlist"`
	}

	err := c.Get(u, &result)
	if err != nil {
		return nil, err
	}
	return result.Users, nil
}
