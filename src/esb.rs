use crate::bridge::{
    usb_protocol::{MessageBuilder, UsbMessage},
    CmdCodes,
};

pub struct EsbMessage {
    pub address: [u8; 5],
    pub id: u8,
    pub err: u8,
    pub payload: Vec<u8>,
}

impl EsbMessage {
    pub fn from_usb_message(usb_msg: UsbMessage) -> EsbMessage {
        EsbMessage {
            address: [1, 2, 3, 4, 5],
            id: usb_msg.payload[6],
            err: usb_msg.payload[7],
            payload: usb_msg.payload,
        }
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
        assert_eq!(usbmsg.err, 0xFF);
        assert_eq!(
            usbmsg.payload.as_slice(),
            [0xde, 0xad, 0xbe, 0xef, 0x00, 0x10, 0xFF, 1, 2, 3, 4, 5, 6]
        );
    }
}
