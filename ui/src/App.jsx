import * as React from 'react'
import { useState, useEffect } from 'react'

// Import Chakra UI components
import {
  Box,
  VStack,
  Heading,
  Text,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
} from '@chakra-ui/react'

// Import custom components
import UserCountDisplay from './components/UserCountDisplay'
import SpotifyConnectForm from './components/SpotifyConnectForm'
import ManualArtistForm from './components/ManualArtistForm'

// Import custom hooks and utilities
import useFormValidation from './hooks/useFormValidation'
import { fetchUserCount } from './utils/api'

function App() {
  // State for user count and limit status
  const [userCount, setUserCount] = useState(0);
  const [maxUsers, setMaxUsers] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);
  const [isLimitReached, setIsLimitReached] = useState(false);
  
  // State for selected artists
  const [selectedArtists, setSelectedArtists] = useState([]);
  
  // Form validation hook
  const {
    formData,
    errors,
    handleChange,
    validateForm,
    resetForm
  } = useFormValidation({
    firstName: '',
    lastName: '',
    email: '',
    phoneNumber: ''
  });
  
  // Function to fetch user count from API
  const getUserCount = async () => {
    try {
      const data = await fetchUserCount();
      
      if (data.count !== undefined) {
        setUserCount(data.count);
      }
      
      // Get the max users from the API response
      if (data.maxUsers !== undefined) {
        setMaxUsers(data.maxUsers);
        // Check if we've reached the limit based on the max_users from the API
        setIsLimitReached(data.count >= data.maxUsers);
      }
    } catch (error) {
      console.error('Error fetching user count:', error);
    }
  };

  // Use effect to load user count on mount and poll every 30 seconds
  useEffect(() => {
    // Fetch user count immediately on mount
    getUserCount();
    
    // Set up polling interval
    const interval = setInterval(getUserCount, 30000); // 30 seconds
    
    // Clean up interval on unmount
    return () => clearInterval(interval);
  }, []);
  
  // Use effect to handle animation when userCount changes
  useEffect(() => {
    if (userCount > 0) {
      setIsAnimating(true);
      const timer = setTimeout(() => {
        setIsAnimating(false);
      }, 500);
      return () => clearTimeout(timer);
    }
  }, [userCount]);
  
  return (
    <Box bg="#faf0e6" minHeight="100vh" display="flex" justifyContent="center" alignItems="center" py={8}>
      <Box p={8} width="60%" maxWidth="700px" minWidth="350px" bg="#fafcff" borderRadius='lg'>
        <VStack spacing={6} align="stretch">
          <UserCountDisplay 
            userCount={userCount}
            maxUsers={maxUsers}
            isLimitReached={isLimitReached}
            isAnimating={isAnimating}
          />
          
          <Heading as="h1" size="xl" textAlign="center">
            Music Match
          </Heading>
          <Text size="s" textAlign="center">
            A tiny application that matches you with your musical soulmate
          </Text>
          
          <Tabs isFitted variant="enclosed" colorScheme="spotifygreen">
            <TabList mb="1em">
              <Tab>Select Artists Manually</Tab>
              <Tab isDisabled>Connect with Spotify (Coming Soon) </Tab>
            </TabList>
            <TabPanels>
              {/* Manual Artist Selection Tab */}
              <TabPanel>
                <ManualArtistForm 
                  formData={formData}
                  errors={errors}
                  handleChange={handleChange}
                  validateForm={validateForm}
                  isLimitReached={isLimitReached}
                  selectedArtists={selectedArtists}
                  setSelectedArtists={setSelectedArtists}
                  fetchUserCount={getUserCount}
                  resetForm={resetForm}
                />
              </TabPanel>
              {/* Spotify Connection Tab */}
              <TabPanel>
                <SpotifyConnectForm 
                  formData={formData}
                  errors={errors}
                  handleChange={handleChange}
                  validateForm={validateForm}
                  isLimitReached={isLimitReached}
                />
              </TabPanel>
            </TabPanels>
          </Tabs>
        </VStack>
      </Box>
    </Box>
  );
}

export default App;
