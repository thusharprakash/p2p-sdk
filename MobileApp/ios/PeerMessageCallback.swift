//
//  PeerMessageCallback.swift
//  MobileApp
//
//  Created by APPLE on 11/06/24.
//

import Foundation

import Foundation
import P2p

public class PeerMessageCallback: NSObject, P2pPeerMessageCallbackProtocol{
  private let callback: (String?) -> Void

  public init(callback: @escaping (String?) -> Void) {
    self.callback = callback
    super.init()
  }
  
  public func onMessage(_ p0: String?) {
    callback(p0)
  }

}
