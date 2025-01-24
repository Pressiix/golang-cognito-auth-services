package services

import (
	"encoding/json"
	"errors"
	"go-api/types"
	"log"
	"os"
	"sync"
)

const filePath = "users.json"

// Mutex to handle concurrent file access
var fileMutex sync.Mutex

var (
    usersCache []types.User // In-memory cache for users
    cacheMutex sync.Mutex   // Mutex for thread-safe access to the cache
)

// InitializeCache loads users from the file into memory
func InitializeCache() error {
    fileMutex.Lock()
    defer fileMutex.Unlock()

    users, err := GetAllUsers()
    if err != nil {
        return err
    }

    cacheMutex.Lock()
    defer cacheMutex.Unlock()
    usersCache = users
    return nil
}


// GetAllUsers retrieves all users from the cache
func GetAllUsers() ([]types.User, error) {
    cacheMutex.Lock()
    defer cacheMutex.Unlock()

    // If the cache is not initialized, load data from the file
    if len(usersCache) == 0 {
        log.Println("Cache is empty. Loading data from the file...")
        
        file, err := os.Open(filePath)
        if err != nil {
            return nil, err
        }
        defer file.Close()

        var users []types.User
        decoder := json.NewDecoder(file)
        if err := decoder.Decode(&users); err != nil {
            return nil, err
        }

        usersCache = users // Populate the cache
    }

    // Return the cached users
    return usersCache, nil
}



// GetUserByID retrieves a user by index (ID) from the cache
func GetUserByID(id int) (types.User, error) {
    cacheMutex.Lock()
    defer cacheMutex.Unlock()

    if id < 0 || id >= len(usersCache) {
        return types.User{}, errors.New("user not found")
    }

    return usersCache[id], nil
}

// CreateUser adds a new user to the cache and JSON file
func CreateUser(user types.User) error {
    cacheMutex.Lock()
    defer cacheMutex.Unlock()

    // Step 1: Append the new user to the cache
    usersCache = append(usersCache, user)

    // Step 2: Write the updated cache to the file
    return writeUsersToFile(usersCache)
}

// UpdateUser updates a user by index (ID) in the cache and JSON file
func UpdateUser(id int, updatedUser types.User) error {
    cacheMutex.Lock()
    defer cacheMutex.Unlock()

    if id < 0 || id >= len(usersCache) {
        return errors.New("user not found")
    }

    usersCache[id] = updatedUser
    return writeUsersToFile(usersCache)
}

// DeleteUser removes a user by index (ID) from the cache and JSON file
func DeleteUser(id int) error {
    cacheMutex.Lock()
    defer cacheMutex.Unlock()

    if id < 0 || id >= len(usersCache) {
        return errors.New("user not found")
    }

    usersCache = append(usersCache[:id], usersCache[id+1:]...)
    return writeUsersToFile(usersCache)
}

// writeUsersToFile writes all users back to the JSON file
func writeUsersToFile(users []types.User) error {
    file, err := os.Create(filePath) // Truncates the file
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ") // Pretty-print JSON
    return encoder.Encode(users)
}

