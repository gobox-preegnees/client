package consts

// This mod is used when the application sends a snapshot while running
const UPDATE_MODE = 200
// This mod is used when the application has just started working
const ON_START_MODE = 100
// This mod is used when the application received at least some error during operation 
// (either when creating a snapshot or when sending it)
const ERROR_MODE = 300