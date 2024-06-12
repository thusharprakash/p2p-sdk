//
//  PeerListCallback.swift
//  MobileApp
//
//  Created by APPLE on 12/06/24.
//

import Foundation

import P2p

public class PeerListCallback: NSObject, P2pPeerCallbackProtocol{
  private let callback: (String?) -> Void

  public init(callback: @escaping (String?) -> Void) {
    self.callback = callback
    super.init()
  }
  
  public func onMessage(_ p0: String?) {
    callback(p0)
  }

}
