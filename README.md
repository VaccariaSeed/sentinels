# sentinels
多规约采集控制器

//todo 数据库采用VictoriaMetrics


## modbus tcp 控制
```go
value := make(map[string]string)
value["startAddr"] = "0x01"
value["length"] = "0x02"
value["value"] = "33,34"
time.Sleep(5 * time.Second)
opt := &model.Operate{
    UniqueIdentifier: "2121212121",
    ReplySize:        0,
    SignType:         "address",
    Sign:             "test_coils",
    SendTime:         0,
    ValidityPeriod:   0,
    Cmd: &model.OperateCmd{
        Timeout:  0,
        CmdType:  "setCmd",
        FuncCode: "0x10",
        Value:    value,
    },
}
exec, err := task.GTP.Exec(opt)
if err != nil {
    fmt.Println(err)
    return
}
fmt.Println("exec result:", exec)
```

## modbus tcp 直接抄读
```go
value := make(map[string]string)
value["startAddr"] = "0x01"
value["length"] = "0x02"
time.Sleep(5 * time.Second)
opt := &model.Operate{
    UniqueIdentifier: "2121212121",
    ReplySize:        0,
    SignType:         "address",
    Sign:             "test_coils",
    SendTime:         0,
    ValidityPeriod:   0,
    Cmd: &model.OperateCmd{
        Timeout:  0,
        CmdType:  "copyRead",
        FuncCode: "0x01",
        Value:    value,
    },
}
exec, err := task.GTP.Exec(opt)
if err != nil {
    fmt.Println(err)
    return
}
fmt.Println("exec result:", exec)
```