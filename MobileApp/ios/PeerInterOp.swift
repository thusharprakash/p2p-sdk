//
//  PeerInterOp.swift
//  MobileApp
//
//  Created by APPLE on 11/06/24.
//

import Foundation
import P2p

@objc public class PeerInterOp:NSObject{
  
  @objc public override init(){
    super.init()
  }
  
  @objc public func start() -> String{
    let config = P2pNodeConfig()
    let documentsDirectory = NSSearchPathForDirectoriesInDomains(.documentDirectory, .userDomainMask, true)[0]
    let dbPath = documentsDirectory.appending("/events.db")
    config?.setStorage(dbPath)
    let id = P2pStartP2PChat(config)
    return id
  }
  
  @objc public func startSDKSubscription(callback: @escaping (String?) -> Void){
    let sdkCallback = PeerMessageCallback{message in
      callback(message)
    }
    P2pStartSubscription(sdkCallback)
  }
  
  
  @objc public func sendMessage(message:String){
    P2pPublishMessage(message,nil)
  }
  
  @objc public func observePeers(callback: @escaping (String?) -> Void){
    let sdkCallback = PeerListCallback{message in
      callback(message)
    }
    P2pSubscribeToPeers(sdkCallback)
  }
}
