pub struct EsbMessage {
    pub address: [u8; 5],
    pub id: u8,
    pub err: u8,
    pub payload: Vec<u8>,
}