/*
 * SPDX-FileCopyrightText: 2024 TokenG Project
 * SPDX-License-Identifier: Apache-2.0
 */

package com.google.gsm.tokenhelper;

import android.accounts.Account;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.EditText;
import android.widget.Spinner;
import android.widget.TextView;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * Main fragment for TokenHelper app.
 * Provides UI to select account, enter password, and fetch tokens via microG.
 */
public class TokenFragment extends Fragment {
    
    private MicroGBridge bridge;
    private ExecutorService executor;
    
    // UI elements
    private Spinner accountSpinner;
    private EditText passwordInput;
    private Button addAccountBtn;
    private Button getPhotosTokenBtn;
    private Button getAuthStringBtn;
    private Button getMasterTokenBtn;
    private TextView resultText;
    
    private Account[] accounts;
    private String lastResult;
    
    @Override
    public void onCreate(@Nullable Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        bridge = new MicroGBridge(requireContext());
        executor = Executors.newSingleThreadExecutor();
    }
    
    @Nullable
    @Override
    public View onCreateView(@NonNull LayoutInflater inflater, @Nullable ViewGroup container,
                             @Nullable Bundle savedInstanceState) {
        View view = inflater.inflate(R.layout.fragment_token, container, false);
        
        accountSpinner = view.findViewById(R.id.account_spinner);
        passwordInput = view.findViewById(R.id.password_input);
        addAccountBtn = view.findViewById(R.id.add_account_btn);
        getPhotosTokenBtn = view.findViewById(R.id.get_photos_token_btn);
        getAuthStringBtn = view.findViewById(R.id.get_auth_string_btn);
        getMasterTokenBtn = view.findViewById(R.id.get_master_token_btn);
        resultText = view.findViewById(R.id.result_text);
        
        setupUI();
        return view;
    }
    
    private void setupUI() {
        // Check if microG is installed
        if (!bridge.isMicroGInstalled()) {
            resultText.setText(R.string.no_microg);
            getPhotosTokenBtn.setEnabled(false);
            getAuthStringBtn.setEnabled(false);
            getMasterTokenBtn.setEnabled(false);
            return;
        }
        
        // Load accounts
        refreshAccounts();
        
        // Add account button
        addAccountBtn.setOnClickListener(v -> {
            startActivity(bridge.getLoginIntent());
        });
        
        // Get Photos Token button
        getPhotosTokenBtn.setOnClickListener(v -> {
            fetchToken("getPhotosToken");
        });
        
        // Get Auth String button
        getAuthStringBtn.setOnClickListener(v -> {
            fetchToken("getPhotosAuthString");
        });
        
        // Get Master Token button
        getMasterTokenBtn.setOnClickListener(v -> {
            fetchToken("getMasterToken");
        });
        
        // Copy result on click
        resultText.setOnClickListener(v -> {
            if (lastResult != null && !lastResult.isEmpty()) {
                copyToClipboard("Token", lastResult);
            }
        });
    }
    
    private void refreshAccounts() {
        accounts = bridge.getAccounts();
        
        if (accounts.length == 0) {
            accountSpinner.setEnabled(false);
            String[] noAccounts = {"No accounts - tap Add Account"};
            accountSpinner.setAdapter(new ArrayAdapter<>(requireContext(),
                    android.R.layout.simple_spinner_dropdown_item, noAccounts));
        } else {
            accountSpinner.setEnabled(true);
            String[] names = new String[accounts.length];
            for (int i = 0; i < accounts.length; i++) {
                names[i] = accounts[i].name;
            }
            accountSpinner.setAdapter(new ArrayAdapter<>(requireContext(),
                    android.R.layout.simple_spinner_dropdown_item, names));
        }
    }
    
    private void fetchToken(String method) {
        if (accounts == null || accounts.length == 0) {
            Toast.makeText(requireContext(), R.string.no_account_selected, Toast.LENGTH_SHORT).show();
            return;
        }
        
        int selectedIndex = accountSpinner.getSelectedItemPosition();
        if (selectedIndex < 0 || selectedIndex >= accounts.length) {
            Toast.makeText(requireContext(), R.string.no_account_selected, Toast.LENGTH_SHORT).show();
            return;
        }
        
        String email = accounts[selectedIndex].name;
        String password = passwordInput.getText().toString().trim();
        
        resultText.setText("Fetching...");
        
        executor.execute(() -> {
            Bundle result;
            switch (method) {
                case "getPhotosToken":
                    result = bridge.getPhotosToken(email, password);
                    break;
                case "getPhotosAuthString":
                    result = bridge.getPhotosAuthString(email, password);
                    break;
                case "getMasterToken":
                    result = bridge.getMasterToken(email, password);
                    break;
                default:
                    result = new Bundle();
                    result.putBoolean("success", false);
                    result.putString("error", "Unknown method");
            }
            
            requireActivity().runOnUiThread(() -> {
                displayResult(result);
            });
        });
    }
    
    private void displayResult(Bundle result) {
        if (result == null) {
            resultText.setText("Error: No response");
            lastResult = null;
            return;
        }
        
        boolean success = result.getBoolean("success", false);
        
        if (success) {
            String token = result.getString("token");
            String authString = result.getString("authString");
            
            lastResult = token != null ? token : authString;
            
            if (lastResult != null) {
                // Truncate for display
                String display = lastResult.length() > 100 
                    ? lastResult.substring(0, 50) + "..." + lastResult.substring(lastResult.length() - 30)
                    : lastResult;
                resultText.setText("✓ Success!\n\n" + display + "\n\n(Tap to copy)");
                Toast.makeText(requireContext(), R.string.token_copied, Toast.LENGTH_SHORT).show();
                copyToClipboard("Token", lastResult);
            } else {
                resultText.setText("✓ Success but no token returned");
            }
        } else {
            String error = result.getString("error", "Unknown error");
            resultText.setText("✗ Error:\n\n" + error);
            lastResult = null;
        }
    }
    
    private void copyToClipboard(String label, String text) {
        ClipboardManager clipboard = (ClipboardManager) 
            requireContext().getSystemService(Context.CLIPBOARD_SERVICE);
        ClipData clip = ClipData.newPlainText(label, text);
        clipboard.setPrimaryClip(clip);
    }
    
    @Override
    public void onResume() {
        super.onResume();
        // Refresh accounts in case user added one
        if (bridge != null && bridge.isMicroGInstalled()) {
            refreshAccounts();
        }
    }
    
    @Override
    public void onDestroy() {
        super.onDestroy();
        if (executor != null) {
            executor.shutdown();
        }
    }
}
