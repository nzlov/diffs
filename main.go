package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("must agrs dirs a and b")
		return
	}

	ma := map[string]string{}
	mb := map[string]string{}
	errs := []error{}
	erCh := make(chan error)

	wait := &sync.WaitGroup{}
	wait.Add(2)

	go func() {
		defer wait.Done()
		if err := loaddir(os.Args[1], ma, erCh); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wait.Done()
		if err := loaddir(os.Args[2], mb, erCh); err != nil {
			panic(err)
		}
	}()
	go func() {
		for err := range erCh {
			errs = append(errs, err)
		}
	}()
	wait.Wait()
	close(erCh)

	fmt.Printf("目录%s共有%d个文件\n", os.Args[1], len(ma))
	fmt.Printf("目录%s共有%d个文件\n", os.Args[2], len(mb))

	mc := map[string]string{}
	for k, v := range ma {
		if mbv, ok := mb[k]; ok {
			delete(ma, k)
			delete(mb, k)
			if v != mbv {
				mc[k] = v + ":" + mbv
			}
		}
	}
	if len(ma) > 0 {
		fmt.Println(os.Args[1] + "多余文件:")
		sortprint(ma, "  ", false)
	}
	if len(mb) > 0 {
		fmt.Println(os.Args[2] + "多余文件:")
		sortprint(mb, "  ", false)
	}
	if len(mc) > 0 {
		fmt.Println("不同文件:")
		sortprint(mc, "  ", true)
	}
	if len(errs) > 0 {
		fmt.Println("错误信息:")
		for _, v := range errs {
			fmt.Println("  " + v.Error())
		}
	}
}

func sortprint(m map[string]string, prefix string, v bool) {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if v {
			fmt.Println(prefix, k, ":", m[k])
		} else {
			fmt.Println(prefix, k)
		}
	}
}

func loaddir(root string, m map[string]string, erCh chan error) error {
	n := len(root)
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		m5, err := hash_file_md5(path)
		fmt.Println(path, ":", m5)
		m[path[n:]] = m5
		if err != nil {
			erCh <- err
		}
		return nil
	})
}

func hash_file_md5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil

}
