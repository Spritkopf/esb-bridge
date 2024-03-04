use crc16;

const SYNC_BYTE: u8 = 0xB4;

/// Representation of a USB protocol message
pub struct Message {
    id: u8,
    err: u8,
    payload: Vec<u8>,
}

impl Message {
    pub fn new(msg_id: u8, payload: Vec<u8>) -> Message {
        Message {
            id: msg_id,
            err: 0,
            payload: payload,
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut data = [vec![SYNC_BYTE, self.id, self.err, self.payload.len() as u8], self.payload.clone()].concat();
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
    fn to_bytes() {
        let msg = Message::new(2, vec![0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16]);
        assert_eq!(msg.to_bytes(), vec![180, 2, 0, 7, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x18, 0x53]);
    }
}
