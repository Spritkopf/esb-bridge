pub mod usb_protocol;

use std::collections::HashMap;
use std::sync::{mpsc, Arc, Mutex};
use std::thread;
use std::time::Duration;
use usb_protocol::{UsbMessage, UsbProtocol};

use crate::esb::EsbMessage;
use log;

/// Default timeout for USB commands that don't involve actual ESB communication
const DEFAULT_TIMEOUT: Duration = Duration::from_millis(200);

pub enum CmdCodes {
    CmdVersion = 0x10,        // Get firmware version
    CmdSetCentralAddr = 0x21, // Set central Pipeline address
    CmdTransfer = 0x30,       // Transfer message
    CmdSend = 0x31,           // Send a message without waiting for a reply
    _CmdTest = 0x61,          // test command, do not use
    CmdIrq = 0x80,            // interrupt callback, only from device to host
    CmdRx = 0x81,             // callback from incoming ESB message
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
                let (tx, rx) = mpsc::channel::<UsbMessage>();
                protocol.add_listener(CmdCodes::CmdRx as u8, tx);

                let listeners: Arc<Mutex<HashMap<u8, mpsc::Sender<EsbMessage>>>> =
                    Arc::new(Mutex::new(HashMap::new()));
                let thread_listeners = listeners.clone();
                thread::spawn(move || loop {
                    if let Ok(incoming_msg) = rx.try_recv() {
                        log::debug!(
                            "Incoming USB Message: {:02X}, {:?}",
                            incoming_msg.id,
                            incoming_msg.payload
                        );
                        let esb_message = EsbMessage::from_usb_message(incoming_msg).unwrap();
                        log::debug!("Incoming ESB Message: ID {:02X} Payload {:?}", esb_message.id, esb_message.payload);
                        let l = thread_listeners.lock().unwrap();
                        match l.get(&esb_message.id) {
                            Some(listen_channel) => {
                                log::debug!("Incoming ESB message for listener with CMD {:02X}", &esb_message.id);
                                listen_channel.send(esb_message).unwrap();
                            }
                            None => log::warn!(
                                "Got ESB message but no listener registered for CMD 0x{:02X}",
                                esb_message.id
                            ),
                        }

                    }
                });

                Ok(Bridge {
                    usb_protocol: protocol,
                    listeners,
                })
            }
            Err(e) => Err(e),
        }
    }

    /// Reads and returns the firmware version of the connected Bridge device
    pub fn get_firmware_version(&mut self) -> Result<String, String> {
        match self.usb_protocol.transfer(
            &UsbMessage::new(CmdCodes::CmdVersion as u8, vec![]).unwrap(),
            Duration::from_millis(500),
        ) {
            Ok(answer) => Ok(String::from(format!(
                "v{}.{}.{}",
                answer.payload[0], answer.payload[1], answer.payload[2]
            ))),
            _ => Err(String::from("Error reading Firmware version")),
        }
    }

    /// Transfer an ESB message and return the answer
    pub fn transfer(&mut self, msg: EsbMessage, timeout: Duration) -> Result<EsbMessage, String> {
        match self.usb_protocol.transfer(&msg, timeout) {
            Ok(answer) => EsbMessage::from_usb_message(answer),
            _ => Err(String::from("Error transmitting message")),
        }
    }

    /// Sends an ESB message
    /// This method does blindly send a message and can't verify its successful transmission
    pub fn send(&mut self, _msg: UsbMessage) -> Result<(), String> {
        Err(String::from("Not implemented yet"))
    }

    /// Set the ESB central address for the connected Bridge device
    /// The central address is the pipeline address the bridge listens on for incoming notifications from peripheral
    /// devices
    /// params:
    /// - addr: 5-byte ESB pipeline address
    pub fn set_central_address(&mut self, addr: &[u8; 5]) -> Result<(), String> {
        let msg = UsbMessage::new(CmdCodes::CmdSetCentralAddr as u8, addr.to_vec()).unwrap();

        match self.usb_protocol.transfer(&msg, DEFAULT_TIMEOUT) {
            Ok(_) => Ok(()),
            _ => Err(String::from("Error setting central address message")),
        }
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

    fn init_logger() {
        let _ = env_logger::builder()
            // Include all events in tests
            .filter_level(log::LevelFilter::max())
            // Ensure events are captured by `cargo test`
            .is_test(true)
            // Ignore errors initializing the logger if tests race to configure it
            .try_init();
    }

    #[test]
    #[ignore]
    fn get_fw() {
        let mut bridge = Bridge::new(String::from("/dev/ttyACM0")).unwrap();

        println!("Version: {}", bridge.get_firmware_version().unwrap());
    }

    /// This test tries to get the firmware version of a ESB peripheral device, testing the basic transmission of ESB messages
    #[test]
    #[ignore]
    fn transfer() {
        let mut bridge = Bridge::new(String::from("/dev/ttyACM0")).unwrap();

        // DUT address 123:45:67.89:01, 0x10 is the GetFWVersion-Command of that device
        let dut_addr: [u8; 5] = [123, 45, 67, 89, 1];
        let msg = EsbMessage::new(dut_addr, 0x10, Vec::new()).unwrap();
        let version = bridge.transfer(msg, Duration::from_millis(300)).unwrap();
        println!("ESB Peripheral FW Version: {:?}", version.payload);
    }

    /// This test registers a listener and waits for a incoming ESB message
    #[test]
    #[ignore]
    fn receive() {
        init_logger();
        let mut bridge = Bridge::new(String::from("/dev/ttyACM0")).unwrap();

        let (tx, rx) = mpsc::channel::<EsbMessage>();
        bridge.add_listener(0x33, tx);

        let rx_msg = rx.recv_timeout(Duration::from_secs(5)).unwrap();

        println!(
            "Test Received: ID {:?}, Payload {:?}",
            rx_msg.id, rx_msg.payload
        );
    }
}
