use crc16;

const PACKET_SIZE: usize = 64;
const HEADER_SIZE: usize = 4;
const CRC_SIZE: usize = 2;
const SYNC_BYTE: u8 = 0xB4;
const MAX_PL_LEN: usize = PACKET_SIZE - HEADER_SIZE - CRC_SIZE;

/// Representation of a USB protocol message
pub struct Message {
    id: u8,
    err: u8,
    payload: Vec<u8>,
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

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut data = [vec![SYNC_BYTE, self.id, self.err, self.payload.len() as u8], 
                        self.payload.clone(), 
                        vec![0; MAX_PL_LEN - self.payload.len()]
                       ].concat();
        let checksum = crc(&data);
        println!("Checksum: {:X}", checksum);
        data.push(checksum as u8);
        data.push((checksum >> 8) as u8);
        data
    }
}

fn crc(data: &[u8]) -> u16 {
    println!("{:X?}", data);
    crc16::State::<crc16::CCITT_FALSE>::calculate(data)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn crc_test() {
        let v: Vec<u8> = vec![1, 2, 3, 4];
        assert_eq!(crc16::State::<crc16::CCITT_FALSE>::calculate(&v[..]), 35267);
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
        assert_eq!(bytes, [vec![180, 2, 0, 7, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16], vec![0x00; 51], vec![0x00, 0x71]].concat());
        
        let msg = Message::new(2, vec![0x00; MAX_PL_LEN]).unwrap();
        let bytes = msg.to_bytes();

        assert_eq!(bytes.len(), PACKET_SIZE);
        assert_eq!(bytes, [vec![180, 2, 0, MAX_PL_LEN as u8], vec![0x00; MAX_PL_LEN], vec![0x4B, 0xE6]].concat());
    }
}
