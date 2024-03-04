use serialport

pub mod usb_protocol;

pub struct Bridge {
    device: String,
    port: u32,
}

impl Bridge {
    pub fn new(device: String, port: u32) -> Result<Bridge, String> {

        // todo: connect to Device here, return error if necessary

        
        
    
        Ok(Bridge { device, port })
        //Err(String::from("could not connnnnect!"))
    }

    pub fn get_firmware_version(&self) -> Result<&str, String> {

        Ok("v0.1.0")
    }

    pub fn transfer(&self) -> Result<(), String> {
        Ok(())
    }
}
