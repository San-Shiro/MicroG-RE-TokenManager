/*
 * TokenG-specific constants that override basement library values.
 * Uses a unique package name to avoid class conflicts.
 */
package org.microg.gms.tokenmanager;

/**
 * TokenG-specific constants. The basement library uses "app.revanced" as
 * BASE_PACKAGE_NAME,
 * but TokenG needs "com.google.gsm.token". Import this class in TokenG's auth
 * code
 * to get the correct values.
 */
public class TokenGConstants {
    // TokenG's base package name
    public static final String BASE_PACKAGE_NAME = "com.google.gsm.token";

    // TokenG's account type (same as BASE_PACKAGE_NAME by microG convention)
    public static final String ACCOUNT_TYPE = "com.google.gsm.token";

    // TokenG's package name (applicationId)
    public static final String GMS_PACKAGE_NAME = "com.google.gsm.token.android.gms";

    // Google's real package name - used for API spoofing (DO NOT CHANGE)
    public static final String GOOGLE_GMS_PACKAGE_NAME = "com.google.android.gms";

    // Google's real signature - used for API spoofing (DO NOT CHANGE)
    public static final String GMS_PACKAGE_SIGNATURE_SHA1 = "38918a453d07199354f8b19af05ec6562ced5788";
}
