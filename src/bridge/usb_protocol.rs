use crc16;
use log;
use std::collections::HashMap;
use std::sync::{mpsc, Arc, Mutex};
use std::time::Duration;
use std::thread;

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
    handle: Option<thread::JoinHandle<()>>,
    listeners: Arc<Mutex<HashMap<u8, mpsc::Sender<Message>>>>,
    tx_channel_sender: mpsc::Sender<Message>,
    rx_channel_receiver: mpsc::Receiver<Message>,
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
        if crc(&bytes[..IDX_CRC]) != (((bytes[IDX_CRC + 1] as u16) << 8) | (bytes[IDX_CRC] as u16))
        {
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
    pub fn new(device: String) -> Result<UsbProtocol, String> {
        let port_result = serialport::new(&device, SERIAL_PORT_BAUDRATE)
            .timeout(Duration::from_millis(SERIAL_PORT_TIMEOUT))
            .open();

        match port_result {
            Ok(mut port) => {
                let listeners: Arc<Mutex<HashMap<u8, mpsc::Sender<Message>>>> = Arc::new(Mutex::new(HashMap::new()));
                let thread_listeners = listeners.clone();

                let (tx_ch_sender, tx_ch_receiver) = mpsc::channel::<Message>();
                let (rx_ch_sender, rx_ch_receiver) = mpsc::channel::<Message>();
                
                // Start listening for incoming messages
                // This function will spawn a thread that runs endlessly
                // The thread will wait for messages on the TX channel (used for direct transfers)
                // Each received USB message will be analyzed; if it is an answer to a transfer message,
                // it is returned on the RX channel. If it matches a registered listener,
                // it will be relayed to that listeners channel instead.
                let handle = Some(thread::spawn(move || {
                    let mut read_buffer: Vec<u8> = vec![0; PACKET_SIZE];

                    struct Answer {
                        cmd_id: u8,
                        pending: bool,
                    }
                    let mut answer: Answer = Answer{
                        cmd_id: 0,
                        pending: false
                    };

                    loop {
                        // see if there is a message available for transfer
                        if !answer.pending {
                            if let Ok(tx_msg) = tx_ch_receiver.try_recv() {
                                let write_buffer = tx_msg.to_bytes();
                                port
                                    .write(&write_buffer) // blocks
                                    .unwrap();
                                // TODO: add error checking
                                log::debug!("Sending message: ID 0x{:02X} Payload {:?}", tx_msg.id, tx_msg.payload);
                                
                                // expect to received a answer
                                answer.pending = true;
                                answer.cmd_id = tx_msg.id;
   
                            }
                        }

                        // check incoming messages
                        let available_bytes = port.bytes_to_read().unwrap();
                        if available_bytes >= PACKET_SIZE as u32 {
                            port.read(&mut read_buffer) // blocks
                                .unwrap();
                            match Message::from_bytes(&read_buffer) {
                                Some(msg) => {
                                    // check if we are waiting for an answer, otherwise check for registered listeners
                                    if answer.pending && (answer.cmd_id == msg.id){
                                        log::debug!(
                                            "Got Answer for message: ID 0x{:02X} Err 0x{:02X} Payload {:?}",
                                            &msg.id, &msg.err, &msg.payload
                                        );
                                        rx_ch_sender.send(msg).unwrap();
                                    } else {
                                        let l = thread_listeners.lock().unwrap();
                                        match l.get(&msg.id) {
                                            Some(listen_channel) => {
                                                log::debug!(
                                                    "Got message for listener with CMD {}",
                                                    &msg.id
                                                );
                                                listen_channel.send(msg).unwrap();
                                            }
                                            None => log::debug!(
                                                "Got message but no listener registered for CMD 0x{:02X}",
                                                msg.id
                                            ),
                                        }
                                    }
                                }
                                None => log::debug!("Discarding invalid message"),
                            }
                        }
                    }
                }));
                Ok(UsbProtocol {
                    handle: handle,
                    listeners: listeners,
                    tx_channel_sender: tx_ch_sender,
                    rx_channel_receiver: rx_ch_receiver,
                })
            }
            Err(_) => Err(String::from(format!(
                "Unable to open serial port {}",
                device
            ))),
        }
    }

    /// Transfers a USB Message and returns the answer
    pub fn transfer(&mut self, msg: Message, timeout: Duration) -> Result<Message, String> {
        
        self.tx_channel_sender.send(msg).unwrap();

        match self.rx_channel_receiver.recv_timeout(timeout) {
            Ok(msg) => Ok(msg),
            Err(err) => Err(String::from(format!(
                "Timeout waiting for an answer ({:?})", err)
            )),
        }
    }

    /// Add a listener for a specific command ID. 
    /// If a listener for the supplied ID is already registered, it is overwritten
    /// Params:
    /// - cmd_id: ID of the message to listen for
    /// - channel: a mpsc channel on which the received message is releayed. 
    pub fn add_listener(&mut self, cmd_id: u8, channel: mpsc::Sender<Message>) {
        let mut listeners = self.listeners.lock().unwrap();

        listeners.insert(cmd_id, channel);
        log::info!("Registered listener for cmd_id 0x{:02X}", cmd_id);
    }
}

#[cfg(test)]
mod tests {
    use crate::bridge::CmdCodes;

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
        let expected_bytes = [
            vec![105, 2, 0, 7, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16],
            vec![0x00; 51],
            vec![0x28, 0xDB],
        ]
        .concat();
        assert_eq!(bytes.len(), PACKET_SIZE);
        assert_eq!(bytes, expected_bytes);

        // Edge case: No padding bytes in payload
        let msg = Message::new(2, vec![0x00; MAX_PL_LEN]).unwrap();
        let bytes = msg.to_bytes();

        assert_eq!(bytes.len(), PACKET_SIZE);
        assert_eq!(
            bytes,
            [
                vec![105, 2, 0, MAX_PL_LEN as u8],
                vec![0x00; MAX_PL_LEN],
                vec![0x63, 0x4C]
            ]
            .concat()
        );
    }
    #[test]
    fn from_bytes() {
        let bytes = [
            vec![105, 2, 3, 7, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16],
            vec![0x00; 51],
            vec![0xD8, 0x06],
        ]
        .concat();
        let msg = Message::from_bytes(&bytes).unwrap();
        assert_eq!(msg.id, 2);
        assert_eq!(msg.err, 3);
        assert_eq!(msg.payload.len(), 7);
        assert_eq!(msg.payload, vec![0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16]);
    }

    #[test]
    #[ignore]
    /// This test needs manual interaction and a connected device. 
    /// Once the test is started the "test" button on the device must be pressed within 3 seconds, 
    /// which sends a message with ID CMD_IRQ to the host
    fn irq() {
        let mut prot = UsbProtocol::new(String::from("/dev/ttyACM0")).unwrap();

        let (rx_sender, rx_receiver) = mpsc::channel::<Message>();

        prot.add_listener(CmdCodes::CmdIrq as u8, rx_sender);

        match rx_receiver.recv_timeout(Duration::from_secs(3)) {
            Ok(received) => println!("Got: {:?}", received.to_bytes()),
            Err(e) => println!("ERROR:  {:?}", e),
        }
    }

    #[test]
    #[ignore]
    /// This test needs a connected device. A normal transfer is tested with Command ID 0x10, which
    /// is the firmware version in case of the ESB-bridge firmware
    fn transfer() {
        let mut prot = UsbProtocol::new(String::from("/dev/ttyACM0")).unwrap();

        let msg = Message::new(0x12, vec![]).unwrap();
        match prot.transfer(msg, Duration::from_millis(500)) {
            Ok(received) => println!("Got Answer, result code: 0x{:02X}", received.err),
            Err(e) => println!("ERROR:  {:?}", e),
        }
    }
}
