use esb_bridge::server::usb_protocol::Message;

fn main() {
    let msg = Message::new(12, vec![1,2,3]).unwrap();
    println!("Hello Message {:?}", msg.payload);
}