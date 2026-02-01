/*
 * SPDX-FileCopyrightText: 2024 TokenG Project
 * SPDX-License-Identifier: Apache-2.0
 */

package com.google.gsm.tokenhelper;

import android.os.Bundle;

import androidx.appcompat.app.AppCompatActivity;

/**
 * Main activity for TokenHelper app.
 */
public class MainActivity extends AppCompatActivity {
    
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        
        if (savedInstanceState == null) {
            getSupportFragmentManager()
                .beginTransaction()
                .replace(R.id.fragment_container, new TokenFragment())
                .commit();
        }
    }
}
