package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	serverchan_sdk "github.com/easychen/serverchan-sdk-golang"
)

type Airdrop struct {
	Token           string `json:"token"`
	Name            string `json:"name"`
	Date            string `json:"date"`
	Time            string `json:"time"`
	Points          string `json:"points"`
	Amount          string `json:"amount"`
	Type            string `json:"type"`
	Phase           int    `json:"phase"`
	Status          string `json:"status"`
	SystemTimestamp int64  `json:"system_timestamp"`
	Completed       bool   `json:"completed"`
	ContractAddress string `json:"contract_address"`
	ChainID         string `json:"chain_id"`
}

type ApiResponse struct {
	Airdrops []Airdrop `json:"airdrops"`
}

// 配置结构体
type Config struct {
	SendKeys []string `json:"sendkeys"`
	Interval int      `json:"interval"`
}

// 读取配置文件
func loadConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// 发送到Server酱
func sendToServerChan(msg string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	title := "今日空投播报"
	for _, sendkey := range cfg.SendKeys {
		resp, err := serverchan_sdk.ScSend(sendkey, title, msg, nil)
		if err != nil {
			fmt.Printf("推送Server酱失败: %v\n", err)
		} else {
			fmt.Println("Server酱响应:", resp)
		}
	}
	return nil
}

func getAirdrop() *ApiResponse {
	url := "https://alpha123.uk/api/data?t=1751632712002&fresh=1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("referer", "https://alpha123.uk/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var apiResp ApiResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		log.Fatal(err)
	}

	for i, item := range apiResp.Airdrops {
		if item.Phase == 2 && item.Date != "" && item.Time != "" {
			// 时间加18小时
			layout := "2006-01-02 15:04"
			parsed, err := time.Parse(layout, item.Date+" "+item.Time)
			if err == nil {
				parsed = parsed.Add(18 * time.Hour)
				item.Date = parsed.Format("2006-01-02")
				item.Time = parsed.Format("15:04")
			}
			apiResp.Airdrops[i] = item
		}
	}

	return &apiResp
}

// 获取token单价
func fetchTokenPrice(token string) (float64, error) {
	url := fmt.Sprintf("https://alpha123.uk/api/price/%s?t=%d&fresh=1", token, time.Now().UnixMilli())
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	// 设置必要的Header
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("referer", "https://alpha123.uk/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Mobile Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var result struct {
		Success bool    `json:"success"`
		Price   float64 `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if !result.Success {
		return 0, fmt.Errorf("price fetch failed")
	}
	return result.Price, nil
}

func getSendMsgAndSnapshot() (string, string) {
	apiResp := getAirdrop()
	msg := "| 项目 | 时间 | 积分 | 数量 | 阶段 | 价格(USD) |\n|---|---|---|---|---|---|\n"
	snapshot := ""
	for i, item := range apiResp.Airdrops {
		amount, err := strconv.Atoi(item.Amount)
		if err != nil {
			fmt.Printf("转换数量失败: %v\n", err)
			amount = 0
		}
		// 比较日期是否是今天
		if item.Date != time.Now().Format("2006-01-02") {
			continue
		}

		price, err := fetchTokenPrice(item.Token)
		if err != nil {
			fmt.Printf("获取%s价格失败: %v\n", item.Token, err)
			price = 0
		}
		msg += fmt.Sprintf("| %s(%s) | %s %s | %s | %s | %d | %.2f |\n",
			item.Token, item.Name, item.Date, item.Time, item.Points, item.Amount, item.Phase, price*float64(amount))
		snapshot += fmt.Sprintf("%s|%s|%s|%s|%s|%d\n",
			item.Token, item.Name, item.Date, item.Time, item.Amount, item.Phase)
		apiResp.Airdrops[i] = item
	}
	return msg, snapshot
}
func hashMsg(msg string) string {
	h := md5.New()
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	var lastHash string
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	interval := time.Duration(cfg.Interval) * time.Minute

	for {
		msg, snapshot := getSendMsgAndSnapshot()
		currentHash := hashMsg(snapshot)

		if currentHash != lastHash && msg != "" {
			fmt.Println("检测到空投信息变化，推送通知...")
			fmt.Println(msg)
			if err := sendToServerChan(msg); err != nil {
				fmt.Println("推送Server酱失败:", err)
			}
			lastHash = currentHash
		} else {
			fmt.Println("无变化，无需推送。")
		}
		time.Sleep(interval)
	}
}
