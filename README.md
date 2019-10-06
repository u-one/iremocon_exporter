# iremocon_exporter

This is prometheus exporter for iRemocon which exports temperature, humidity, brightness.

## usage
```
go build
 ./iremocon_exporter -ir_host 192.168.1.XX
```

then check `http://localhost:8080/metrics`

you can also change exporter port
```
 ./iremocon_exporter -ir_host 192.168.1.XX -ex_port 8083
```
