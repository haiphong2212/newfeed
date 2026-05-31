@echo off
REM Create auth-service directory structure
mkdir d:\newfeed\services\auth-service\internal\domain 2>nul
mkdir d:\newfeed\services\auth-service\internal\repository 2>nul
mkdir d:\newfeed\services\auth-service\internal\infrastructure 2>nul
mkdir d:\newfeed\services\auth-service\internal\usecase 2>nul
mkdir d:\newfeed\services\auth-service\internal\config 2>nul
mkdir d:\newfeed\services\auth-service\internal\delivery\grpc 2>nul
mkdir d:\newfeed\services\auth-service\internal\delivery\http 2>nul
mkdir d:\newfeed\services\auth-service\proto 2>nul
echo Auth-service directory structure created successfully!
