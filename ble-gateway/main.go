package main

import (
	"fmt"
	"log"
	"time"
	"tinygo.org/x/bluetooth"
)

// BLE 어댑터 선언
var adapter = bluetooth.DefaultAdapter

// UUID를 처리하여 올바른 형식으로 출력하는 함수 (엔디안 변환 제거)
func formatUUID(uuid bluetooth.UUID) string {
	return uuid.String()  // UUID를 그대로 출력
}

func main() {
	// BLE 어댑터 활성화
	fmt.Println("Initializing BLE adapter...")
	must("enable BLE stack", adapter.Enable())

	for {
		// 일정 시간 동안 스캔 (5초 동안 스캔)
		fmt.Println("Scanning for BLE devices...")

		err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			if result.LocalName() == "" {
				return // 이름이 없는 장치는 무시
			}

			deviceAddress := result.Address.String()

			fmt.Printf("Found device: %s, RSSI: %d, Address: %s\n", result.LocalName(), result.RSSI, deviceAddress)

			// 연결 시도
			fmt.Println("Attempting to connect to device...")
			device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
			if err != nil {
				log.Printf("Failed to connect to device: %v", err)
				time.Sleep(5 * time.Second)
				return
			}
			defer device.Disconnect()

			// 서비스 UUID 출력 (엔디안 변환 제거 후 그대로 출력)
			services, err := device.DiscoverServices(nil)
			if err != nil {
				log.Printf("Failed to discover services: %v", err)
				return
			}

			fmt.Println("Service UUIDs for the device:")
			for _, service := range services {
				// UUID를 그대로 출력
				fmt.Printf("Service UUID: %v\n", formatUUID(service.UUID()))
			}
		})
		must("start scan", err)

		// 5초 동안 스캔
		time.Sleep(5 * time.Second)
		fmt.Println("Stopping scan...")

		// 10초 동안 대기 후 다시 스캔
		time.Sleep(10 * time.Second)
	}
}

// 에러 핸들링 헬퍼 함수
func must(action string, err error) {
	if err != nil {
		log.Fatalf("Failed to %s: %v", action, err)
	}
}
