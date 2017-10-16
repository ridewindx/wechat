package mch

import (
	"encoding/json"
)

func SceneInfo(shopID, shopName, shopAreaCode, shopAddress string) string {
	if len(shopID) > 32 || len(shopName) > 64 || len(shopAreaCode) > 6 || len(shopAddress) > 128 {
		panic("invalid scene info arguments")
	}
	bytes, err := json.Marshal(map[string]map[string]string{
		"store_info": {
			"id":        shopID,
			"name":      shopName,
			"area_code": shopAreaCode,
			"address":   shopAddress,
		},
	})
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func IOSAppSceneInfo(appName, bundleID string) string {
	bytes, err := json.Marshal(map[string]map[string]string{
		"h5_info": {
			"type":      "IOS",
			"app_name":  appName,
			"bundle_id": bundleID,
		},
	})
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func AndroidAppSceneInfo(appName, packageName string) string {
	bytes, err := json.Marshal(map[string]map[string]string{
		"h5_info": {
			"type":         "Android",
			"app_name":     appName,
			"package_name": packageName,
		},
	})
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func H5SceneInfo(wapName, wapURL string) string {
	bytes, err := json.Marshal(map[string]map[string]string{
		"h5_info": {
			"type":     "Wap",
			"wap_name": wapName,
			"wap_url":  wapURL,
		},
	})
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
