#include <Arduino.h>
#include <BLEDevice.h>
#include <BLEUtils.h>
#include <BLEServer.h>

#define SERVICE_UUID        "123e4567-e89b-12d3-a456-426614174000"
#define CHARACTERISTIC_UUID "123e4567-e89b-12d3-a456-426614174001"

// Variable to track the client connection state
bool deviceConnected = false;

// Define the behavior for client connection/disconnection by inheriting BLEServerCallbacks
class MyServerCallbacks: public BLEServerCallbacks {
  void onConnect(BLEServer* pServer) {
    deviceConnected = true; // Set to true when the client is connected
    Serial.println("Client connected!"); // Print message when connected
  }

  void onDisconnect(BLEServer* pServer) {
    deviceConnected = false; // Set to false when the client is disconnected
    Serial.println("Client disconnected!"); // Print message when disconnected
  }
};

void setup() {
  Serial.begin(9600);
  Serial.println("Starting BLE work!");

  // Initialize the BLE device and create a server
  BLEDevice::init("balogin_user1");  // Set device name
  BLEServer *pServer = BLEDevice::createServer();

  // Set callbacks for client connection/disconnection events
  pServer->setCallbacks(new MyServerCallbacks());

  // Create a BLE service
  BLEService *pService = pServer->createService(SERVICE_UUID);

  // Create a BLE characteristic (optional)
  BLECharacteristic *pCharacteristic = pService->createCharacteristic(
                                         CHARACTERISTIC_UUID,
                                         BLECharacteristic::PROPERTY_READ |
                                         BLECharacteristic::PROPERTY_WRITE
                                       );
  pCharacteristic->setValue("Hello World"); // Set initial value for the characteristic
  pService->start(); // Start the service

  // Start advertising
  BLEAdvertising *pAdvertising = BLEDevice::getAdvertising();
  pAdvertising->addServiceUUID(SERVICE_UUID); // Include the service UUID in the advertisement
  pAdvertising->setScanResponse(true);  // Enable scan response
  pAdvertising->setMinPreferred(0x06);  // Set to fix iPhone connection issues
  pAdvertising->setMinPreferred(0x12);
  BLEDevice::startAdvertising();  // Start advertising
  Serial.println("Advertising started!");
}

void loop() {
  if (deviceConnected) {
    // Additional actions when the client is connected
  }
  delay(2000);
}
