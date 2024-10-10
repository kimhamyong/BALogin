package main

import (
    "fmt"
    "log"
    "sync"
    "time"
    "tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

// 이미 연결된 장치의 UUID를 저장하는 맵과 락
var connectedDevices = make(map[string]bool)
var mu sync.Mutex

// RSSI 임계값
const RSSIThreshold = -90

// UUID를 처리하여 올바른 형식으로 출력하는 함수
func formatUUID(uuid bluetooth.UUID) string {
    return uuid.String()
}

// BLE 장치 연결이 끊겼을 때 처리하는 함수
func handleDisconnect(deviceUUID string) {
    mu.Lock()
    defer mu.Unlock()
    if connectedDevices[deviceUUID] {
        fmt.Printf("Device %s disconnected.\n", deviceUUID)
        connectedDevices[deviceUUID] = false
    }
}

// BLE 장치가 연결되었을 때 처리하는 함수
func handleConnect(deviceUUID string) {
    mu.Lock()
    defer mu.Unlock()
    if !connectedDevices[deviceUUID] {
        fmt.Printf("Device %s connected.\n", deviceUUID)
        connectedDevices[deviceUUID] = true
    }
}

// BLE 스캔 재시작 및 최신 장치 상태 반영 함수
func restartScan() {
    for {
        time.Sleep(10 * time.Second)

        fmt.Println("Restarting BLE scan to refresh device states...")

        err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
            if result.LocalName() == "" {
                return
            }

            deviceUUID := result.Address.String()

            if result.RSSI <= RSSIThreshold {
                handleDisconnect(deviceUUID)
                return
            }

            device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
            if err != nil {
                handleDisconnect(deviceUUID)
                return
            }
            defer device.Disconnect()

            services, err := device.DiscoverServices(nil)
            if err != nil {
                handleDisconnect(deviceUUID)
                return
            }

            for _, service := range services {
                uuid := formatUUID(service.UUID())
                if uuid != "00001801-0000-1000-8000-00805f9b34fb" {
                    if result.RSSI > RSSIThreshold {
                        handleConnect(deviceUUID)
                        fmt.Printf("UUID: %s, RSSI: %d\n", uuid, result.RSSI) // UUID 출력
                    } else {
                        handleDisconnect(deviceUUID)
                    }
                }
            }
        })

        if err != nil {
            log.Printf("Error restarting BLE scan: %v", err)
        }
    }
}

func main() {
    fmt.Println("Initializing BLE adapter...")
    must("enable BLE stack", adapter.Enable())

    go restartScan()

    select {}
}

func must(action string, err error) {
    if err != nil {
        log.Fatalf("Failed to %s: %v", action, err)
    }
}
