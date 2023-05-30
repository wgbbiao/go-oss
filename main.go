package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	v2 "gopkg.in/yaml.v2"
)

type Config struct {
	BaseDir string `yaml:"base_dir"`
	Oss     struct {
		Endpoint        string `yaml:"endpoint"`
		AccessKeyId     string `yaml:"accessKeyId"`
		AccessKeySecret string `yaml:"accessKeySecret"`
	}
}

func main() {
	// 加载ymal配置文件
	data, err := os.ReadFile("./config.yaml")

	conf := new(Config)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	} else {
		out, err := v2.Marshal(conf)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		// 保存配置文件
		err = os.WriteFile("./config.yaml", out, 0777)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
	}
	if err := v2.Unmarshal(data, conf); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	if conf.Oss.Endpoint == "" || conf.Oss.AccessKeyId == "" || conf.Oss.AccessKeySecret == "" {
		fmt.Println("Error: oss config is empty.")
		os.Exit(-1)
	}

	fmt.Println(conf)
	// 创建 OSS 客户端实例
	client, err := oss.New(conf.Oss.Endpoint, conf.Oss.AccessKeyId, conf.Oss.AccessKeySecret)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// 获取存储空间
	bucketName := "pm-zjk-01"
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// 遍历存储空间中的所有文件
	marker := oss.Marker("")
	for {
		// 列举文件
		lsRes, err := bucket.ListObjects(oss.MaxKeys(1000), marker)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}

		// 下载文件
		for _, object := range lsRes.Objects {
			// 构造本地存储路径
			localPath := path.Join(conf.BaseDir, object.Key)
			fmt.Println(localPath)
			// 判断目录是否存在
			if err := os.MkdirAll(conf.BaseDir, 0777); err != nil {
				fmt.Println("Error:", err)
				os.Exit(-1)
			}

			// 判断文件是否存在
			if _, err := os.Stat(localPath); err == nil {
				fmt.Println(object.Key + " already exists.")
				continue
			}

			// 创建文件夹
			if err := os.MkdirAll("./"+path.Dir(localPath), 0777); err != nil {
				fmt.Println("Error:", err)
				os.Exit(-1)
			}

			if strings.HasSuffix(object.Key, "/") {
				fmt.Println(object.Key + " is a folder.")
				continue
			}
			// 下载文件
			err = bucket.GetObjectToFile(object.Key, localPath)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(-1)
			}
			fmt.Println(object.Key + " downloaded.")
		}

		// 判断是否还有未遍历的文件
		if !lsRes.IsTruncated {
			break
		}
		marker = oss.Marker(lsRes.NextMarker)
	}

	fmt.Println("All files downloaded.")
}
