mod usb_protocol;

use std::collections::HashMap;
use std::sync::{mpsc, Arc, Mutex};
use std::time::Duration;
use usb_protocol::{UsbProtocol,Message};

use crate::esb::EsbMessage;


enum CmdCodes {
    CmdVersion     = 0x10,   // Get firmware version
    CmdTransfer    = 0x30,   // Transfer message
    CmdSend        = 0x31,   // Send a message without waiting for a reply
    _CmdTest        = 0x61,   // test command, do not use
    _CmdIrq         = 0x80,   // interrupt callback, only from device to host
    CmdRx          = 0x81,   // callback from incoming ESB message
}

pub struct Bridge {
    usb_protocol: UsbProtocol,
    listeners: Arc<Mutex<HashMap<u8, mpsc::Sender<EsbMessage>>>>,
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
                    listeners: Arc::new(Mutex::new(HashMap::new()))
                })
            },
            Err(e) => Err(e)
        }
    }

    /// Reads and returns the firmware version of the connected Bridge device
    pub fn get_firmware_version(&mut self) -> Result<String, String> {

        let res = self.usb_protocol.transfer(&Message::new(CmdCodes::CmdVersion as u8, vec![]).unwrap(), Duration::from_millis(500));
        match res {
            Ok(answer) => Ok(String::from(format!("v{}.{}.{}", answer.payload[0], answer.payload[1], answer.payload[2]))),
            _ => Err(String::from("Error reading Firmware version"))
        }      
    }

    /// Transfer an ESB message and return the answer
    fn transfer(&mut self, msg: EsbMessage, timeout: Duration) -> Result<Message, String> {

        Err(String::from("Not implemented yet"))
       
    }

    /// Sends an ESB message
    /// This method does blindly send a message and can't verify its successful transmission
    fn send(&mut self, msg: Message) -> Result<(), String> {

       Err(String::from("Not implemented yet"))
    }

    /// Register a listener for specific type ESB message
    /// If a listener for the supplied ID is already registered, it is overwritten
    /// Params:
    /// - cmd_id: ID of the ESB message to listen for
    /// - channel: a mpsc channel on which the received message is releayed. 
    pub fn add_listener(&mut self, cmd_id: u8, channel: mpsc::Sender<EsbMessage>) {

        let mut listeners = self.listeners.lock().unwrap();

        listeners.insert(cmd_id, channel);
        log::info!("Registered listener for ESB cmd_id {:02X}", cmd_id);
    }

    
}

#[cfg(test)]
mod tests {
    use super::*;

   
    #[test]
    fn get_fw() {
        let mut bridge = Bridge::new(String::from("/dev/ttyACM0")).unwrap();

        println!("Version: {}", bridge.get_firmware_version().unwrap());
    }
}
