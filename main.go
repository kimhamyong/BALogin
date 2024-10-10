package main

import (
    "fmt"
    "log"
    "sync"
    "time"
    "tinygo.org/x/bluetooth"
)

// BLE 어댑터 선언
var adapter = bluetooth.DefaultAdapter

// 이미 연결된 장치의 UUID를 저장하는 맵과 락
var connectedDevices = make(map[string]bool) // 장치의 연결 상태를 저장
var mu sync.Mutex                            // 멀티 고루틴에서 안전하게 맵을 접근하기 위한 락

// RSSI 임계값 (신호 세기가 이 값 이하로 떨어지면 연결이 끊겼다고 간주)
const RSSIThreshold = -90

// UUID를 처리하여 올바른 형식으로 출력하는 함수
func formatUUID(uuid bluetooth.UUID) string {
    return uuid.String() // UUID를 그대로 출력
}

// BLE 장치 연결이 끊겼을 때 처리하는 함수
func handleDisconnect(deviceUUID string) {
    mu.Lock()
    defer mu.Unlock()
    if connectedDevices[deviceUUID] { // 연결 상태가 true인 경우에만 끊김 출력
        fmt.Printf("Device %s disconnected.\n\n", deviceUUID)
        connectedDevices[deviceUUID] = false // 연결 상태 해제
    }
}

// BLE 장치가 연결되었을 때 처리하는 함수
func handleConnect(deviceUUID string) {
    mu.Lock()
    defer mu.Unlock()
    if !connectedDevices[deviceUUID] { // 이전에 연결되지 않은 경우에만 출력
        fmt.Printf("Device %s connected.\n\n", deviceUUID)
        connectedDevices[deviceUUID] = true // 연결 상태를 true로 변경
    }
}

// BLE 스캔 재시작 및 최신 장치 상태 반영 함수
func restartScan() {
    for {
        time.Sleep(10 * time.Second) // 10초마다 BLE 스캔 재시작

        fmt.Println("Restarting BLE scan to refresh device states...")

        err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
            // 이름이 없는 기기는 무시
            if result.LocalName() == "" {
                return
            }

            // 장치의 서비스 UUID를 통해 연결 상태 관리
            deviceUUID := result.Address.String()

            // 연결 시도
            device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
            if err != nil {
                // 연결 실패 시에도 무조건 "disconnected" 출력
                handleDisconnect(deviceUUID) // 연결 실패 시 "disconnected" 처리
                time.Sleep(5 * time.Second)  // 재시도 대기 시간 추가
                return
            }
            defer device.Disconnect()

            // 서비스 UUID를 통해 장치 연결 상태 추적
            services, err := device.DiscoverServices(nil)
            if err != nil {
                handleDisconnect(deviceUUID) // 서비스 탐색 실패 시에도 "disconnected" 처리
                device.Disconnect()
                return
            }

            // UUID 출력 및 연결 상태 관리
            for _, service := range services {
                uuid := formatUUID(service.UUID())
                if uuid != "00001801-0000-1000-8000-00805f9b34fb" {
                    // 사용자 정의 서비스 UUID만 연결 상태 관리
                    if result.RSSI > RSSIThreshold {
                        handleConnect(uuid) // UUID로 연결 관리
                    } else {
                        handleDisconnect(uuid) // UUID로 연결 해제 관리
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
    // BLE 어댑터 활성화
    fmt.Println("Initializing BLE adapter...")
    must("enable BLE stack", adapter.Enable())

    // BLE 스캔 재시작 고루틴 실행
    go restartScan()

    select {} // 메인 함수가 종료되지 않도록 대기
}

// 에러 핸들링 헬퍼 함수
func must(action string, err error) {
    if err != nil {
        log.Fatalf("Failed to %s: %v", action, err)
    }
}

