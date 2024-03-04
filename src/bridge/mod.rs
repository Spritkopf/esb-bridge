pub mod usb_protocol;

use std::time::Duration;

use usb_protocol::{UsbProtocol,Message};
use serialport::{SerialPort, UsbPortInfo};


const SERIAL_PORT_BAUDRATE: u32 = 9600;
const SERIAL_PORT_TIMEOUT: u64 = 100;

enum CmdCodes {
    CmdVersion     = 0x10,   // Get firmware version
    CmdTransfer    = 0x30,   // Transfer message
    CmdSend        = 0x31,   // Send a message
    CmdTest        = 0x61,   // test command, do not use
    CmdIrq         = 0x80,   // interrupt callback, only from device to host
    CmdRx          = 0x81,   // callback from incoming ESB message
}

pub struct Bridge {
    usb_protocol: UsbProtocol,
    port: u32,
}

impl Bridge {
    pub fn new(device: String, port: u32) -> Result<Bridge, String> {

        Ok(Bridge {
            usb_protocol: UsbProtocol::new(device)?, 
            port: port
        })
    }

    pub fn get_firmware_version(&mut self) -> Result<String, String> {

        let res = self.usb_protocol.transfer(Message::new(CmdCodes::CmdVersion as u8, vec![]).unwrap());
        match res {
            Ok(answer) => Ok(String::from(format!("v{}.{}.{}", answer.payload[0], answer.payload[1], answer.payload[2]))),
            _ => Err(String::from("Error reading Firmware version"))
        }      
    }

    /// Transfer a message
    fn transfer(&mut self, msg: Message) -> Result<Message, String> {

       Err(String::from("Not implemented yet"))
    }
}
