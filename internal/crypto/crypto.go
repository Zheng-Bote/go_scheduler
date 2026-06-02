/**
 * SPDX-FileComment: Crypto
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file crypto.go
 * @brief AES-256-GCM encryption/decryption with Argon2id key derivation
 * @version 1.0.0
 * @date 2026-06-02
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

// Package crypto provides AES-256-GCM encryption and decryption primitives
// protected by Argon2id key derivation. It is used by the config package to
// securely store database credentials and admin tokens on disk, and by the
// encrypt-config CLI tool.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters as per requirements
const (
	ArgonTime      = 3
	ArgonMemory    = 64 * 1024 // 64 MB
	ArgonThreads   = 1
	ArgonKeyLength = 32
)

// DeriveKey derives a 32-byte key from a password and salt using Argon2id
func DeriveKey(password []byte, salt []byte) []byte {
	return argon2.IDKey(password, salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLength)
}

// Encrypt encrypts data using AES-256-GCM
func Encrypt(plaintext []byte, password []byte) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Result format: salt (16) + nonce (12) + ciphertext
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	
	result := make([]byte, len(salt)+len(nonce)+len(ciphertext))
	copy(result[0:16], salt)
	copy(result[16:16+len(nonce)], nonce)
	copy(result[16+len(nonce):], ciphertext)

	return result, nil
}

// Decrypt decrypts data using AES-256-GCM
func Decrypt(data []byte, password []byte) ([]byte, error) {
	if len(data) < 16+12 {
		return nil, errors.New("invalid encrypted data size")
	}

	salt := data[:16]
	nonce := data[16 : 16+12]
	ciphertext := data[16+12:]

	key := DeriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}
