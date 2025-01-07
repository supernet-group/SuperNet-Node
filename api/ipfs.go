package api

import (
	"SuperNet-Node/config"
	"SuperNet-Node/utils"
	logs "SuperNet-Node/utils/log_utils"
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type UploadCidItem struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
}

func UploadFileToIPFS(ipfsNodeUrl, filePath string, timeout time.Duration) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("> os.Open: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return "", fmt.Errorf("> writer.CreateFormFile: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("> io.Copy: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("> writer.Close: %v", err)
	}

	req, err := http.NewRequest("POST", ipfsNodeUrl+"/rpc/api/v0/add?stream-channels=true&progress=false", body)
	if err != nil {
		return "", fmt.Errorf("> http.NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("> client.Do: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("> io.ReadAll: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(respBody))
	for scanner.Scan() {
		line := scanner.Text()
		var item UploadCidItem
		err := json.Unmarshal([]byte(line), &item)
		if err != nil {
			return "", fmt.Errorf("> json.Unmarshal: %v", err)
		}
		return item.Hash, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("> scanner.Err: %v", err)
	}
	return "", fmt.Errorf("no lines in response")
}

func CopyFileInIPFS(ipfsNodeUrl, source, destination string) error {
	req, err := http.NewRequest("POST", ipfsNodeUrl+"/rpc/api/v0/files/cp?parents=true&arg="+source+"&arg="+destination, nil)
	if err != nil {
		return fmt.Errorf("> http.NewRequest: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("> client.Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("> io.ReadAll: %v", err)
		}
		return fmt.Errorf("> unexpected status code: %v, boby: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func RmFileInIPFS(ipfsNodeUrl, destination string) error {
	req, err := http.NewRequest("POST", ipfsNodeUrl+"/rpc/api/v0/files/rm?arg="+destination+"&recursive=true&force=true", nil)
	if err != nil {
		return fmt.Errorf("> http.NewRequest: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("> client.Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("> io.ReadAll: %v", err)
		}
		return fmt.Errorf("> unexpected status code: %v, boby: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func FileLsInIPFS(ipfsNodeUrl string) error {
	req, err := http.NewRequest("POST", ipfsNodeUrl+"/rpc/api/v0/files/ls", nil)
	if err != nil {
		return fmt.Errorf("> http.NewRequest: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("> client.Do: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("> io.ReadAll: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("> unexpected status code: %v, boby: %s", resp.StatusCode, string(respBody))
	}
	logs.Normal(fmt.Sprintf("boby: %s", string(respBody)))

	return nil
}

type ResLs struct {
	Objects []struct {
		Hash  string `json:"Hash"`
		Links []struct {
			Name   string  `json:"Name"`
			Hash   string  `json:"Hash"`
			Size   float64 `json:"Size"`
			Type   uint16  `json:"Type"`
			Target string  `json:"Target"`
		} `json:"Links"`
	} `json:"Objects"`
}

func LsInIPFS(ipfsNodeUrl, cid string) (ResLs, error) {
	var resLs ResLs
	req, err := http.NewRequest("POST", ipfsNodeUrl+"/rpc/api/v0/ls?arg="+cid, nil)
	if err != nil {
		return resLs, fmt.Errorf("> http.NewRequest: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return resLs, fmt.Errorf("> client.Do: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resLs, fmt.Errorf("> io.ReadAll: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return resLs, fmt.Errorf("> unexpected status code: %v, boby: %s", resp.StatusCode, string(respBody))
	}
	logs.Normal(fmt.Sprintf("IPFS ls boby: %s", string(respBody)))

	err = json.Unmarshal(respBody, &resLs)
	if err != nil {
		return resLs, fmt.Errorf("> json.Unmarshal: %v", err)
	}

	if resLs.Objects[0].Hash != cid {
		return resLs, fmt.Errorf("> unexpected cid: %v , target cid: %v", resLs.Objects[0].Hash, cid)
	}
	return resLs, nil
}

func GetFileInIPFS(ipfsNodeUrl, path, target string) error {
	req, err := http.NewRequest("POST", ipfsNodeUrl+"/rpc/api/v0/get?arg="+path+"&output="+target, nil)
	if err != nil {
		return fmt.Errorf("> http.NewRequest: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("> client.Do: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("> io.ReadAll: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("> unexpected status code: %v, boby: %s", resp.StatusCode, string(respBody))
	}
	logs.Normal(fmt.Sprintf("boby: %s", string(respBody)))

	return nil
}

type ResFileStat struct {
	Hash           string  `json:"Hash"`
	Size           uint64  `json:"Size"`
	CumulativeSize float64 `json:"CumulativeSize"`
	Blocks         uint64  `json:"Blocks"`
	Type           string  `json:"Type"`
}

func FileStatInIPFS(ipfsNodeUrl, destination string) (ResFileStat, error) {
	var resFileStat ResFileStat

	req, err := http.NewRequest("POST", ipfsNodeUrl+"/rpc/api/v0/files/stat?arg="+destination, nil)
	if err != nil {
		return resFileStat, fmt.Errorf("> http.NewRequest: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return resFileStat, fmt.Errorf("> client.Do: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resFileStat, fmt.Errorf("> io.ReadAll: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return resFileStat, fmt.Errorf("> unexpected status code: %v, boby: %s", resp.StatusCode, string(respBody))
	}

	err = json.Unmarshal(respBody, &resFileStat)
	if err != nil {
		return resFileStat, fmt.Errorf("> json.Unmarshal: %v , boby: %s", err, string(respBody))

	}
	return resFileStat, nil
}

func DownloadProjectInIPFS(ipfsNodeUrl, dest, cid string) error {
	resLs, err := LsInIPFS(ipfsNodeUrl, cid)
	if err != nil {
		return fmt.Errorf("> LsInIPFS: %v", err)
	}

	for _, link := range resLs.Objects[0].Links {
		if link.Type == 1 {
			newDest := dest + "/" + link.Name
			utils.EnsureDirExists(newDest)
			err = DownloadProjectInIPFS(ipfsNodeUrl, newDest, link.Hash)
			if err != nil {
				return fmt.Errorf("> DownloadProjectInIPFS recursion: %v", err)
			}
		} else if link.Type == 2 {
			var modelURL []utils.DownloadURL
			modelURL = append(modelURL, utils.DownloadURL{
				URL:      config.GlobalConfig.Console.IpfsNodeUrl + "/ipfs" + utils.EnsureLeadingSlash(link.Hash),
				Checksum: "",
				Name:     link.Name,
			})

			err = utils.DownloadFiles(dest, modelURL)
			if err != nil {
				return fmt.Errorf("> DownloadFiles: %v", err)
			}
		} else {
			logs.Warning(fmt.Sprintf("link: %v", link))
		}
	}
	return nil
}
