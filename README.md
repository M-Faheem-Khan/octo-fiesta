This repository contains code that I have written to better understand the inner workings of a VPN by implementing one in Golang natively.

## Stages
1. Create a Tunnel/Tap interface (Implemented)
2. Read/Send data(packets)   
3. Encapsulate Packet - Custom Payload
    1. UDP  
    2. TCP  
4. Encrypt Traffic  
    1. Key Exchange  
    2. Authenticated Encryption(AEAD)  
5. Implement RFCs  
    1. Wireguard  
    2. IPSec  
    3. DTLS 1.3  

<!-- EOF -->
