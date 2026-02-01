/*
 * SPDX-FileCopyrightText: 2024 TokenG Project
 * SPDX-License-Identifier: Apache-2.0
 */

package com.google.gsm.tokenhelper;

import android.accounts.Account;
import android.accounts.AccountManager;
import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.net.Uri;
import android.os.Bundle;

/**
 * Bridge class to communicate with microG's TokenManagerProvider.
 * This allows TokenHelper to fetch tokens via microG without duplicating auth
 * code.
 */
public class MicroGBridge {

    // microG package name (ReVanced variant)
    private static final String MICROG_PACKAGE = "app.revanced.android.gms";

    // TokenManagerProvider authority (uses basePackageName)
    private static final String TOKEN_PROVIDER_AUTHORITY = "app.revanced.gms.tokenmanager";

    // Account type used by microG (same as basePackageName)
    private static final String ACCOUNT_TYPE = "app.revanced";

    // Login activity
    private static final String LOGIN_ACTIVITY = "org.microg.gms.auth.login.LoginActivity";

    private final Context context;

    public MicroGBridge(Context context) {
        this.context = context.getApplicationContext();
    }

    /**
     * Check if microG is installed.
     */
    public boolean isMicroGInstalled() {
        try {
            context.getPackageManager().getPackageInfo(MICROG_PACKAGE, 0);
            return true;
        } catch (PackageManager.NameNotFoundException e) {
            return false;
        }
    }

    /**
     * Get all accounts from microG.
     */
    public Account[] getAccounts() {
        AccountManager am = AccountManager.get(context);
        return am.getAccountsByType(ACCOUNT_TYPE);
    }

    /**
     * Launch microG login activity to add an account.
     */
    public Intent getLoginIntent() {
        Intent intent = new Intent();
        intent.setComponent(new ComponentName(MICROG_PACKAGE, LOGIN_ACTIVITY));
        return intent;
    }

    /**
     * Fetch Google Photos token for an account.
     * 
     * @param email    Account email
     * @param password API password (if password auth is enabled in microG)
     * @return Bundle with 'success' boolean and 'token' or 'error' string
     */
    public Bundle getPhotosToken(String email, String password) {
        return callProvider("getPhotosToken", email, password, null, null);
    }

    /**
     * Get Google Photos auth request string.
     */
    public Bundle getPhotosAuthString(String email, String password) {
        return callProvider("getPhotosAuthString", email, password, null, null);
    }

    /**
     * Get master token for an account.
     */
    public Bundle getMasterToken(String email, String password) {
        return callProvider("getMasterToken", email, password, null, null);
    }

    /**
     * Get custom token for any app/scope.
     */
    public Bundle getCustomToken(String email, String password, String packageName, String scope) {
        return callProvider("getCustomToken", email, password, packageName, scope);
    }

    /**
     * Call the TokenManagerProvider.
     */
    private Bundle callProvider(String method, String email, String password,
            String packageName, String scope) {
        try {
            Uri uri = Uri.parse("content://" + TOKEN_PROVIDER_AUTHORITY);

            Bundle extras = new Bundle();
            if (password != null && !password.isEmpty()) {
                extras.putString("password", password);
            }
            if (packageName != null) {
                extras.putString("packageName", packageName);
            }
            if (scope != null) {
                extras.putString("scope", scope);
            }

            Bundle result = context.getContentResolver().call(uri, method, email, extras);

            if (result == null) {
                Bundle error = new Bundle();
                error.putBoolean("success", false);
                error.putString("error", "microG did not respond. Is it installed?");
                return error;
            }

            return result;

        } catch (Exception e) {
            Bundle error = new Bundle();
            error.putBoolean("success", false);
            error.putString("error", "Error: " + e.getMessage());
            return error;
        }
    }
}
