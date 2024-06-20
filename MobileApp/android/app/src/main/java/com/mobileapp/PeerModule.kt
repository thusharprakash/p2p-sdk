package com.mobileapp

import android.os.Build
import android.util.Log
import com.facebook.react.bridge.Arguments
import com.facebook.react.bridge.ReactApplicationContext
import com.facebook.react.bridge.ReactContextBaseJavaModule
import com.facebook.react.bridge.ReactMethod
import com.facebook.react.modules.core.DeviceEventManagerModule
import p2p.P2p


interface PermissionCallback {
    fun onPermitted()
}

class PeerModule(private val reactContext: ReactApplicationContext) : ReactContextBaseJavaModule(reactContext), PermissionCallback {

    override fun getName(): String {
        return "PeerModule"
    }

    @ReactMethod
    fun start() {
        currentActivity?.let {
            if (it is MainActivity) {
                it.permissionCallback = this
            }
            if (PermissionManager(it).requestNetworkPermission()){
                startSdk()
            }
        }
    }

    @ReactMethod
    fun sendMessage(message: String) {
        P2p.publishMessage(message)
    }
    
    @ReactMethod
    fun getLogs():String {
       return P2p.pullLogs()
    }


    private fun sendMessageBackToReact(message: String,tag:String){
        val map = Arguments.createMap()
        map.putString("message", message)
        reactContext
            .getJSModule(
                DeviceEventManagerModule.RCTDeviceEventEmitter::class.java
            )
            .emit(tag, map)
    }

    @ReactMethod
    fun addListener(type: String?) {
        // Keep: Required for RN built in Event Emitter Calls.
    }

    @ReactMethod
    fun removeListeners(type: Int?) {
        // Keep: Required for RN built in Event Emitter Calls.
    }


    private fun startSdk(){
        val path = reactContext.filesDir.absolutePath +"/events.db"
        val mdnsLockerDriver = MDNSLockerDriver(reactContext)
        Thread {
            val nodeConfig = P2p.newNodeConfig()
            nodeConfig.setStorage(path)
            nodeConfig.setNickName(Build.BRAND+"_"+Build.DEVICE)

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                val inet = NetDriver()
                nodeConfig.setNetDriver(inet)
            }
            nodeConfig.setMDNSLocker(mdnsLockerDriver)
            val peerID = P2p.startP2PChat(nodeConfig)
            sendMessageBackToReact(peerID,"PEER_ID")
            P2p.startSubscription {
                sendMessageBackToReact(it,"P2P")
            }
            P2p.subscribeToPeers {
                Log.e("PEERS",it)
                sendMessageBackToReact(it,"PEERS")
            }
        }.start()
    }

    override fun onPermitted() {
       startSdk()
    }
}