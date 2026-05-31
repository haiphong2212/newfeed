#!/usr/bin/env node
/**
 * Script to create auth-service directory structure and files
 * Run: node setup_auth_service.js
 */

const fs = require('fs');
const path = require('path');

const BASE_PATH = path.normalize(r'd:\newfeed\services\auth-service'.replace(/\r/g, ''));

// Define directory structure
const DIRECTORIES = [
    BASE_PATH,
    path.join(BASE_PATH, 'internal'),
    path.join(BASE_PATH, 'internal', 'domain'),
    path.join(BASE_PATH, 'internal', 'repository'),
    path.join(BASE_PATH, 'internal', 'infrastructure'),
    path.join(BASE_PATH, 'internal', 'usecase'),
    path.join(BASE_PATH, 'internal', 'config'),
    path.join(BASE_PATH, 'internal', 'delivery'),
    path.join(BASE_PATH, 'internal', 'delivery', 'grpc'),
    path.join(BASE_PATH, 'internal', 'delivery', 'http'),
    path.join(BASE_PATH, 'proto'),
];

// Create directories
console.log('Creating directories...');
DIRECTORIES.forEach(directory => {
    if (!fs.existsSync(directory)) {
        fs.mkdirSync(directory, { recursive: true });
    }
    console.log(`✓ Created: ${directory}`);
});

console.log('\n✓ Auth-service directory structure created successfully!');
console.log(`✓ Base path: ${BASE_PATH}`);
