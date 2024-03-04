use std::time::Duration;
use std::sync::mpsc;
use serialport::SerialPort;
use crc16;
use log;

const PACKET_SIZE: usize = 64;
const HEADER_SIZE: usize = 4;
const CRC_SIZE: usize = 2;
const SYNC_BYTE: u8 = 0x69;
const MAX_PL_LEN: usize = PACKET_SIZE - HEADER_SIZE - CRC_SIZE;

const IDX_SYNC: usize = 0;
const IDX_ID: usize = 1;
const IDX_ERR: usize = 2;
const IDX_PL_LEN: usize = 3;
const IDX_PL: usize = 4;
const IDX_CRC: usize = PACKET_SIZE - CRC_SIZE;

const SERIAL_PORT_BAUDRATE: u32 = 9600;
const SERIAL_PORT_TIMEOUT: u64 = 100;

/// Representation of a USB protocol message
pub struct Message {
    pub id: u8,
    pub err: u8,
    pub payload: Vec<u8>,
}

pub struct UsbProtocol {
    serial_port: Box<dyn SerialPort>,
    listeners: Vec<Listener>,
}

struct Listener {
    cmd_id: u8,
    channel: mpsc::Sender<Message>,
}

impl Message {
    pub fn new(msg_id: u8, payload: Vec<u8>) -> Result<Message, &'static str> {
        if payload.len() > MAX_PL_LEN {
            return Err("Payload too large");
        }
        Ok(Message {
            id: msg_id,
            err: 0,
            payload: payload,
        })
    }

    /// Convert the message to a vector of bytes which will be sent to the device
    pub fn to_bytes(&self) -> Vec<u8> {
        let mut data = [
            vec![SYNC_BYTE, self.id, self.err, self.payload.len() as u8],
            self.payload.clone(),
            vec![0; MAX_PL_LEN - self.payload.len()],
        ]
        .concat();
        let checksum = crc(&data);
        log::debug!("Calculated checksum: {:X}", checksum);
        data.push(checksum as u8);
        data.push((checksum >> 8) as u8);
        data
    }

    /// Construct a message from a byte slice received from the device
    pub fn from_bytes(bytes: &[u8]) -> Option<Message> {
        // Check Packet size
        if bytes.len() != PACKET_SIZE {
            return None;
        }
        // Check SYNC byte
        if bytes[IDX_SYNC] != SYNC_BYTE {
            return None;
        }

        // Check CRC
        if crc(&bytes[..IDX_CRC]) != (((bytes[IDX_CRC+1] as u16) << 8) | (bytes[IDX_CRC] as u16)) {
            return None;
        }

        let payload_len = bytes[IDX_PL_LEN] as usize;

        Some(Message {
            id: bytes[IDX_ID],
            err: bytes[IDX_ERR],
            payload: bytes[IDX_PL..(IDX_PL + payload_len)].to_vec(),
        })
    }
}

fn crc(data: &[u8]) -> u16 {
    crc16::State::<crc16::CCITT_FALSE>::calculate(data)
}

impl UsbProtocol {
    /// Creates a new UsbProtocol instance and returns it
    pub fn new(device: String) -> Result<Self, String> {

        let port_result = serialport::new(&device, SERIAL_PORT_BAUDRATE)
            .timeout(Duration::from_millis(SERIAL_PORT_TIMEOUT))
            .open();

        match port_result {
            Ok(port) => Ok(
                            UsbProtocol {
                                serial_port: port,
                                listeners: Vec::new(),
                            }
                        ),
            Err(_) => Err(String::from(format!("Unable to open serial port {}", device)))
        }
    }

    /// Transfers a USB Message and returns the answer
    pub fn transfer(&mut self, msg: Message) -> Result<Message, String> {
        let write_buffer = msg.to_bytes();

        // Write to serial port
        let num_tx = self.serial_port
            .write(&write_buffer) // blocks
            .unwrap();
        
        log::debug!("Written bytes: {:?}", num_tx);

        let mut read_buffer: Vec<u8> = vec![0; PACKET_SIZE];

        let n = self.serial_port
            .read(&mut read_buffer) // blocks
            .unwrap();

        log::debug!("Read bytes: {:?} {:?}", n, &read_buffer[..n]);
        
        let answer_msg = Message::from_bytes(&read_buffer);

        match answer_msg {
            Some(msg) => Ok(msg),
            None => Err(String::from(format!("Got no valid answer for request with ID {:?}", msg.id)))
        }
    }

    pub fn add_listener(&mut self, cmd_id: u8, channel: mpsc::Sender<Message>) {
        self.listeners.append(&mut vec![Listener{cmd_id, channel}]);
    }

}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn crc_test() {
        let v: Vec<u8> = vec![1, 2, 3, 4];
        assert_eq!(crc(&v[..]), 35267);
    }

    #[test]
    fn create_msg() {
        let msg = Message::new(2, vec![0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16]).unwrap();

        assert_eq!(msg.id, 2);
        assert_eq!(msg.err, 0);
        assert_eq!(msg.payload, vec![0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16]);
    }

    #[test]
    fn create_msg_err() {
        assert!(Message::new(2, vec![0; 59]).is_err());
    }

    #[test]
    fn to_bytes() {
        let msg = Message::new(2, vec![0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16]).unwrap();
        let bytes = msg.to_bytes();

        assert_eq!(bytes.len(), PACKET_SIZE);
        assert_eq!(
            bytes,
            [
                vec![180, 2, 0, 7, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16],
                vec![0x00; 51],
                vec![0x00, 0x71]
            ]
            .concat()
        );

        // Edge case: No padding bytes in payload
        let msg = Message::new(2, vec![0x00; MAX_PL_LEN]).unwrap();
        let bytes = msg.to_bytes();

        assert_eq!(bytes.len(), PACKET_SIZE);
        assert_eq!(
            bytes,
            [
                vec![180, 2, 0, MAX_PL_LEN as u8],
                vec![0x00; MAX_PL_LEN],
                vec![0x4B, 0xE6]
            ]
            .concat()
        );
    }
    #[test]
    fn from_bytes() {
        let bytes = [
            vec![180, 2, 3, 7, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16],
            vec![0x00; 51],
            vec![0xF0, 0xAC],
        ]
        .concat();
        let msg = Message::from_bytes(&bytes).unwrap();
        assert_eq!(msg.id, 2);
        assert_eq!(msg.err, 3);
        assert_eq!(msg.payload.len(), 7);
        assert_eq!(msg.payload, vec![0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16]);
    }
}
