pub mod usbprotocol;

pub struct Server {
    device: String,
    port: u32,
}

impl Server {
    pub fn new(device: String, port: u32) -> Result<Server, String> {

        // todo: connect to Device here, return error if necessary

        Ok(Server { device, port })
    }

    pub fn start() -> Result<(), String> {

        // todo: start RPC server here (grpc? protobuf? tbd...)
        
        Ok(())
    }
}
