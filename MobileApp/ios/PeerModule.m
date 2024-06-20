//
//  PeerModule.m
//  MobileApp
//
//  Created by APPLE on 11/06/24.
//

#import "PeerModule.h"
#import <React/RCTLog.h>
#import "MobileApp-Swift.h"

@implementation PeerModule
{
  PeerInterOp *peer;
}

// To export a module named IPFSModule
RCT_EXPORT_MODULE(PeerModule);

- (instancetype)init
{
  self = [super init];
  if (self) {
    peer = [[PeerInterOp alloc] init];
  }
  return self;
}

RCT_EXPORT_METHOD(start)
{
  NSString *sdkVersion = @"1.0.0"; // replace with actual SDK version
  RCTLogInfo(@"Starting SDK  %@", sdkVersion);
  NSString *result = [peer start];
  
  [self sendEventWithName:@"PEER_ID" body:@{@"message": result}];
  
  RCTLogInfo(@"Starting subscription");
  [peer startSDKSubscriptionWithCallback:^(NSString * _Nullable message) {
    [self sendEventWithName:@"P2P" body:@{@"message": message}];
  }];
  
  RCTLogInfo(@"Starting peer listing");
  [peer observePeersWithCallback:^(NSString * _Nullable message) {
    [self sendEventWithName:@"PEERS" body:@{@"message": message}];
  }];
}

RCT_EXPORT_METHOD(sendMessage:(NSString *)message)
{
  RCTLogInfo(@"Sending message %@", message);
  [peer sendMessageWithMessage:message];

}

RCT_EXPORT_BLOCKING_SYNCHRONOUS_METHOD(getLogs)
{
  return [peer pullLogs];
}


- (NSArray<NSString *> *)supportedEvents
{
  return @[@"P2P", @"PEER_ID",@"PEERS"];
}

@end
