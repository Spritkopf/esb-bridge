pub mod usb_protocol;
use std::sync::mpsc;
use std::time::Duration;

use usb_protocol::{UsbProtocol,Message};


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
    listeners: Vec<Listener>,
}

impl Bridge {
    /// Returns a Bridge object 
    ///
    /// Creating the Bridge will connect to the USB device and register the default listener
    /// for Command ID "CmdRx" which is used by the ESB-Bridge device to relay received ESB messages
    /// to the host.
    /// 
    /// # Arguments
    ///
    /// * `device` - A string holding the serial port device to open, e.g. "/dev/ttyACM0"
    pub fn new(device: String) -> Result<Bridge, String> {

        match UsbProtocol::new(device) {
            Ok(mut protocol) => {
                let (tx, _rx) = mpsc::channel::<Message>();
                protocol.add_listener(CmdCodes::CmdRx as u8, tx);

                Ok(Bridge {
                    usb_protocol: protocol,
                    listeners: Vec::new(),
                })
            },
            Err(e) => Err(e)
        }
    }

    pub fn get_firmware_version(&mut self) -> Result<String, String> {

        let res = self.usb_protocol.transfer(Message::new(CmdCodes::CmdVersion as u8, vec![]).unwrap(), Duration::from_millis(500));
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
