package com.mobileapp

import android.Manifest
import android.app.Activity
import android.content.pm.PackageManager
import androidx.activity.ComponentActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat

class PermissionManager(private val activity: Activity) {

    fun requestNetworkPermission():Boolean {
        val permissions = arrayOf(Manifest.permission.ACCESS_NETWORK_STATE, Manifest.permission.ACCESS_FINE_LOCATION,
            Manifest.permission.ACCESS_COARSE_LOCATION,
            Manifest.permission.ACCESS_WIFI_STATE,
            Manifest.permission.CHANGE_WIFI_MULTICAST_STATE)
            val neededPermissions = ArrayList<String>()

            for (permission in permissions) {
                if (ContextCompat.checkSelfPermission(activity, permission) != PackageManager.PERMISSION_GRANTED) {
                    neededPermissions.add(permission)
                }
            }

        return if (neededPermissions.isNotEmpty()) {
            ActivityCompat.requestPermissions(activity, neededPermissions.toTypedArray(),
                REQUEST_NETWORK_PERMISSION_CODE)
            false
        }else{
            true
        }
    }

    companion object {
         const val REQUEST_NETWORK_PERMISSION_CODE = 123
    }
}