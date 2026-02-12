# TokenG ProGuard Rules
# Keep all microg and GMS classes (they use reflection heavily)
-dontwarn com.android.org.conscrypt.SSLParametersImpl
-dontwarn org.apache.harmony.xnet.provider.jsse.SSLParametersImpl
-dontwarn java.awt.**
-dontwarn javax.annotation.**
-dontwarn okio.**
-dontwarn com.squareup.okhttp.**
-dontwarn org.slf4j.**
-dontnote

# Keep dynamically loaded GMS classes
-keep public class com.google.android.gms.common.security.ProviderInstallerImpl { public *; }
-keep public class com.google.android.gms.dynamic.IObjectWrapper { public *; }
-keep public class com.google.android.gms.chimera.container.DynamiteLoaderImpl { public *; }
-keep public class com.google.android.gms.dynamite.descriptors.** { public *; }

# Keep AutoSafeParcelables
-keep public class * extends org.microg.safeparcel.AutoSafeParcelable {
    @org.microg.safeparcel.SafeParceled *;
    @org.microg.safeparcel.SafeParcelable.Field *;
    <init>(...);
}

# Keep form data
-keepclassmembers class * {
    @org.microg.gms.common.HttpFormClient$* *;
}

# Keep our stuff
-keep class org.microg.** { *; }
-keep class com.google.android.gms.** { *; }

# Keep asInterface method cause it's accessed from SafeParcel
-keepattributes InnerClasses
-keepclassmembers interface * extends android.os.IInterface {
    public static class *;
}
-keep public class * extends android.os.Binder { public static *; }

# Keep library info
-keep class **.BuildConfig { *; }

# Keep protobuf class builders
-keep public class com.squareup.wire.Message
-keep public class * extends com.squareup.wire.Message
-keep public class * extends com.squareup.wire.Message$Builder { public <init>(...); }

# Keep Android components declared in manifest
-keep public class * extends android.app.Activity
-keep public class * extends android.app.Service
-keep public class * extends android.content.BroadcastReceiver
-keep public class * extends android.content.ContentProvider

# Keep data binding classes
-keep class **.databinding.** { *; }

# Keep Volley
-keep class com.android.volley.** { *; }

# Keep WebView JS interfaces
-keepclassmembers class * {
    @android.webkit.JavascriptInterface <methods>;
}
