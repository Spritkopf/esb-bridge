pub mod usb_protocol;

use std::time::Duration;

use usb_protocol::Message;
use serialport::SerialPort;

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
    device: String,
    port: u32,
    serial_port: Box<dyn SerialPort>
}

impl Bridge {
    pub fn new(device: String, port: u32) -> Bridge {

        Bridge { 
            device: device.clone(),
            port: port, 
            serial_port: serialport::new(device, SERIAL_PORT_BAUDRATE)
                .timeout(Duration::from_millis(SERIAL_PORT_TIMEOUT))
                .open()
                .unwrap()
            }
    }

    pub fn connect(&self) -> Result<(), String>{
        println!("Connecting to device {}", self.device);
        
        Ok(())
    }

    pub fn get_firmware_version(&mut self) -> Result<String, String> {

        let res = self.transfer(Message::new(CmdCodes::CmdVersion as u8, vec![]).unwrap());
        match res {
            Ok(answer) => Ok(String::from(format!("v{}.{}.{}", answer.payload[0], answer.payload[1], answer.payload[2]))),
            _ => Err(String::from("Error reading Firmware version"))
        }

       
    }

    /// Transfer a message
    fn transfer(&mut self, msg: Message) -> Result<Message, String> {

        let write_buffer = msg.to_bytes();

        // Write to serial port
        let num_tx = self.serial_port
            .write(&write_buffer) // blocks
            .unwrap();
        
        //println!("Written bytes: {:?}", num_tx);

        let mut read_buffer: Vec<u8> = vec![0; 64];

        let n = self.serial_port
            .read(&mut read_buffer) // blocks
            .unwrap();

        // For debugging comment these two lines in
        //println!("Read bytes: {:?}", n);
        //println!("{:?}", &read_buffer[..n]);   
        
        let answer_msg = Message::from_bytes(&read_buffer);

        match answer_msg {
            Some(msg) => Ok(msg),
            None => Err(String::from(format!("Got no valid answer for request with ID {:?}", msg.id)))
        }
    }
}
