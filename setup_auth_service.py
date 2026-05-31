#!/usr/bin/env python3
"""
Script to create auth-service directory structure and files
Run: python3 setup_auth_service.py
"""

import os
import sys
from pathlib import Path

BASE_PATH = r"d:\newfeed\services\auth-service"

# Define directory structure
DIRECTORIES = [
    BASE_PATH,
    os.path.join(BASE_PATH, "internal"),
    os.path.join(BASE_PATH, "internal", "domain"),
    os.path.join(BASE_PATH, "internal", "repository"),
    os.path.join(BASE_PATH, "internal", "infrastructure"),
    os.path.join(BASE_PATH, "internal", "usecase"),
    os.path.join(BASE_PATH, "internal", "config"),
    os.path.join(BASE_PATH, "internal", "delivery"),
    os.path.join(BASE_PATH, "internal", "delivery", "grpc"),
    os.path.join(BASE_PATH, "internal", "delivery", "http"),
    os.path.join(BASE_PATH, "proto"),
]

# Create directories
print("Creating directories...")
for directory in DIRECTORIES:
    Path(directory).mkdir(parents=True, exist_ok=True)
    print(f"✓ Created: {directory}")

print("\n✓ Auth-service directory structure created successfully!")
print(f"✓ Base path: {BASE_PATH}")
