# 檔案結構
1. 採模組設計，先在本地目錄執行 go mod init dbx, 這樣整個套件稱之為 dbx，後面會看到好處
1. 將用到的其他模組分別放，使用上因為模組設計，使用 import ""dbx/database" 就可以引入
1. Examples/ 下都是些範例，使用時複製到本目錄的 main.go, 這樣就可以 go build 產生 dbx 執行檔
1. 目前 Examples/ 有幾個範例:
    1. create-table: 最簡單的，產生表格
	1. map: Insert() + MapInsert() 的運用
1. 如果要在 struct 與 map 互相轉換，請見[mapstructure](#2)
1. 編譯與執行:  
  go build && ./dbx

# 參考
[1] [https://pkg.go.dev/reflect](https://pkg.go.dev/reflect)
[2] [https://github.com/mitchellh/mapstructure](https://github.com/mitchellh/mapstructure)
