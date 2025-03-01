use crate::bridge::{
    usb_protocol::{MessageBuilder, UsbMessage},
    CmdCodes,
};

const ESB_PIPE_ADDR_SIZE: usize = 5; // 5-byte pipeline address
const ESB_PACKET_SIZE: usize = 32;
const ESB_HEADER_SIZE: usize = 2 + ESB_PIPE_ADDR_SIZE;
const ESB_MAX_PL_LEN: usize = ESB_PACKET_SIZE - ESB_HEADER_SIZE;

pub struct EsbMessage {
    pub address: [u8; 5],
    pub id: u8,
    pub err: u8,
    pub payload: Vec<u8>,
}

impl EsbMessage {
    pub fn new(
        target_addr: [u8; ESB_PIPE_ADDR_SIZE],
        msg_id: u8,
        payload: Vec<u8>,
    ) -> Result<EsbMessage, String> {
        if payload.len() > ESB_MAX_PL_LEN {
            return Err(String::from("Payload too large"));
        }

        Ok(EsbMessage {
            address: target_addr,
            id: msg_id,
            err: 0,
            payload,
        })
    }

    pub fn from_usb_message(usb_msg: UsbMessage) -> Result<EsbMessage, String> {
        let mut addr = [0; ESB_PIPE_ADDR_SIZE];
        
        if usb_msg.payload.len() < ESB_HEADER_SIZE {
            return Err(String::from(format!("Can't build EsbMessage: Malformed USB packet")));
        }

        let esb_payload_len = usb_msg.payload.len() - ESB_HEADER_SIZE;

        let mut esb_payload = vec![0; esb_payload_len];
        addr.copy_from_slice(
            &usb_msg.payload[ESB_HEADER_SIZE - ESB_PIPE_ADDR_SIZE..ESB_HEADER_SIZE],
        );
        esb_payload.copy_from_slice(&usb_msg.payload[ESB_HEADER_SIZE..]);

        Ok(EsbMessage {
            address: addr,
            id: usb_msg.payload[0],
            err: usb_msg.payload[1],
            payload: esb_payload,
        })
    }
}

impl MessageBuilder for EsbMessage {
    fn build_message(&self) -> UsbMessage {
        let payload = [
            self.address.clone().to_vec(),
            vec![self.id, self.err],
            self.payload.clone(),
        ]
        .concat();
        UsbMessage {
            id: CmdCodes::CmdTransfer as u8,
            err: 0,
            payload,
        }
    }
}

#[cfg(test)]
mod tests {

    use super::*;

    #[test]
    fn build_usb_message() {
        let msg = EsbMessage {
            address: [0xde, 0xad, 0xbe, 0xef, 0x00],
            id: 0x10,
            err: 0xFF,
            payload: vec![1, 2, 3, 4, 5, 6],
        };

        let usbmsg = msg.build_message();

        assert_eq!(usbmsg.id, 0x30); // Transfer code
        assert_eq!(usbmsg.err, 0x00);
        assert_eq!(
            usbmsg.payload.as_slice(),
            [0xde, 0xad, 0xbe, 0xef, 0x00, 0x10, 0xFF, 1, 2, 3, 4, 5, 6]
        );
    }

    #[test]
    fn build_usb_message_no_payload() {
        let msg = EsbMessage {
            address: [0xde, 0xad, 0xbe, 0xef, 0x00],
            id: 0x10,
            err: 0xFF,
            payload: Vec::new(),
        };

        let usbmsg = msg.build_message();

        assert_eq!(usbmsg.id, 0x30); // Transfer code
        assert_eq!(usbmsg.err, 0x00);
        assert_eq!(
            usbmsg.payload.as_slice(),
            [0xde, 0xad, 0xbe, 0xef, 0x00, 0x10, 0xFF]
        );
    }

    #[test]
    fn from_usb_message() {
        let msg = UsbMessage::new(
            30,
            vec![
                0xAA, 0x12, 0xde, 0xad, 0xbe, 0xef, 0x01, 0x10, 0x00, 0x11, 0x00,
            ],
        )
        .expect("error building usb message");
        let esb_msg = EsbMessage::from_usb_message(msg).unwrap();

        assert_eq!(esb_msg.id, 0xAA);
        assert_eq!(esb_msg.err, 0x12);
        assert_eq!(esb_msg.address, [0xde, 0xad, 0xbe, 0xef, 0x01]);
        assert_eq!(esb_msg.payload, [0x10, 0x00, 0x11, 0x00]);
    }

    #[test]
    fn from_usb_message_malformed_packet() {
        let msg = UsbMessage::new(
            30,
            vec![
                0xAA, 0x12, 0xde, 0xad    /* payload does not contain complete ESB package */
            ],
        )
            .expect("error building usb message");
        let result = EsbMessage::from_usb_message(msg);
        assert_eq!(result.err(), Some(String::from("Can't build EsbMessage: Malformed USB packet")));
    }
}
