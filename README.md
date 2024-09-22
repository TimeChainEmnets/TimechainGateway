# TimechainGateway

## cmd/main.go
- 程序启动的入口
  - 首先通过config.Load()载入相关配置信息
  - Load函数中将config.json中的内容解析为Config结构体,执行Decode时通过结构体中定义的json注释进行字段匹配
      ```
        decoder := json.NewDecoder(file)
        err = decoder.Decode(&config)
      ```
  - mqtt.New()创建mqtt broker
  - 创建tcp listener并将其添加到创建的mqtt broker中，用于监听mqtt client