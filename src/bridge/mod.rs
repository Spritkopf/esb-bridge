pub mod usb_protocol;

use std::time::Duration;

use serialport::SerialPort;

const SERIAL_PORT_BAUDRATE: u32 = 9600;
const SERIAL_PORT_TIMEOUT: u64 = 100;

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

    pub fn get_firmware_version(&self) -> Result<String, String> {

        Ok(String::from("v0.1.0")
    }

    pub fn transfer(&mut self) -> Result<(), String> {

        /// TEMPORARY TEST ROUTINE to test two-way communication
        /// To be removed
        let mut write_buffer: Vec<u8> = vec![0; 256];
        write_buffer[0] = b'H';
        write_buffer[1] = b'e';
        write_buffer[2] = b'l';
        write_buffer[3] = b'l';
        write_buffer[4] = b'o';
        write_buffer[5] = b'\n';
        
        let n = 6; // How many bytes to write to serial port.
        
        // Write to serial port
        self.serial_port
            .write(&write_buffer[..n]) // blocks
            .unwrap();

    let mut read_buffer: Vec<u8> = vec![0; 64];

    let n = self.serial_port
        .read(&mut read_buffer) // blocks
        .unwrap();

    println!("{:?}", &read_buffer[..n]);   
    let s1: String = String::from_utf8(read_buffer).unwrap();
    println!("{:?}", s1);   

        Ok(())
    }
}
