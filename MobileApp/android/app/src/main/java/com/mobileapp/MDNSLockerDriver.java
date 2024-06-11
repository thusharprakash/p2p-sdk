package com.mobileapp;

import android.content.Context;
import android.net.wifi.WifiManager;
import android.net.wifi.WifiManager.MulticastLock;

import p2p.NativeMDNSLockerDriver;


public class MDNSLockerDriver implements NativeMDNSLockerDriver {
    private final Context context;
    private MulticastLock multicastLock = null;

    public MDNSLockerDriver(Context context) {
        this.context = context;
    }

    public void lock() {
        WifiManager wifi = (WifiManager) context.getApplicationContext().getSystemService(Context.WIFI_SERVICE);
        multicastLock = wifi.createMulticastLock("multicastLock");
        multicastLock.setReferenceCounted(true);
        multicastLock.acquire();
    }


    public void unlock() {
        if (multicastLock != null) {
            multicastLock.release();
            multicastLock = null;
        }
    }
}