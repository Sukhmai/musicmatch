import * as React from 'react'
import { useState, useEffect } from 'react'
import { MdMusicNote } from "react-icons/md";

// Import Chakra UI components
import {
  Box,
  Text,
  Button,
  FormControl,
  FormLabel,
  Input,
  VStack,
  Heading,
  FormErrorMessage,
  useToast
} from '@chakra-ui/react'

function App() {
  // State for user count and limit status
  const [userCount, setUserCount] = useState(0);
  const [maxUsers, setMaxUsers] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);
  const [isLimitReached, setIsLimitReached] = useState(false);
  
  // Function to fetch user count from API
  const fetchUserCount = async () => {
    try {
      const response = await fetch('/api/spotify.v1.SpotifyService/GetUserCount', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({}), // Empty request as per the proto definition
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data = await response.json();
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
    fetchUserCount();
    
    // Set up polling interval
    const interval = setInterval(fetchUserCount, 30000); // 30 seconds
    
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
  
  // State for form fields
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    phoneNumber: ''
  });
  
  // State for form validation
  const [errors, setErrors] = useState({});
  
  // Toast for form submission feedback
  const toast = useToast();

  // Handle input changes
  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value
    });
  };

  // Validate form
  const validateForm = () => {
    const newErrors = {};
    
    // Validate first name
    if (!formData.firstName.trim()) {
      newErrors.firstName = 'First name is required';
    }
    
    // Validate last name
    if (!formData.lastName.trim()) {
      newErrors.lastName = 'Last name is required';
    }
    
    // Validate email
    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid';
    }
    
    // Validate phone number (optional but must be valid if provided)
    if (formData.phoneNumber && !/^\d{10}$/.test(formData.phoneNumber.replace(/\D/g, ''))) {
      newErrors.phoneNumber = 'Phone number must be 10 digits';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Check if the user limit is reached
    if (isLimitReached) {
      toast({
        title: 'Maximum Users Reached',
        description: 'We have reached the maximum number of users for this round. Please wait for the next round.',
        status: 'warning',
        duration: 5000,
        isClosable: true,
      });
      return;
    }
    
    if (validateForm()) {
      try {
        // Show loading toast
        toast({
          title: 'Connecting to Spotify',
          description: 'Redirecting to Spotify authorization...',
          status: 'info',
          duration: 3000,
          isClosable: true,
        });
        
        // Call the backend to get the Spotify auth URL using the proxied endpoint
        const response = await fetch('/api/spotify.v1.SpotifyService/GetAuthURL', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({}), // Empty request as per the proto definition
        });
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        
        // Store form data in localStorage to retrieve after auth
        localStorage.setItem('userFormData', JSON.stringify(formData));
        
        // Redirect to the Spotify authorization URL
        window.location.href = data.url;
      } catch (error) {
        console.error('Error connecting to Spotify:', error);
        
        // Show error toast
        toast({
          title: 'Connection Error',
          description: 'Failed to connect to Spotify. Please try again.',
          status: 'error',
          duration: 5000,
          isClosable: true,
        });
      }
    }
  };

  return (
    <Box bg="#faf0e6" height="100vh" display="flex" justifyContent="center" alignItems="center">
      <Box p={8} width="60%" maxWidth="700px" minWidth="350px" bg="#fafcff" borderRadius='lg'>
        <VStack spacing={6} align="stretch">
          <Box display="flex" justifyContent="center" flexDirection="column" alignItems="center">
            <Text 
              fontSize="2xl" 
              textAlign="center" 
              bg={isLimitReached ? "#ffe6e6" : "#faf0e6"}
              color={isLimitReached ? "red.600" : "inherit"}
              borderRadius='lg'
              px={4}
              py={2}
              display="inline-block"
              width="fit-content"
              transform={isAnimating ? 'scale(1.1)' : 'scale(1)'}
              transition="transform 0.3s ease-in-out"
            >
              {isLimitReached 
                ? "Maximum Users Reached" 
                : `${maxUsers > 0 ? maxUsers - userCount : '...'} Users until Match Day`}
            </Text>
            {isLimitReached && (
              <Text 
                color="red.600" 
                mt={2} 
                textAlign="center"
                fontSize="md"
              >
                Please wait for the next round to submit your music preferences.
              </Text>
            )}
          </Box>
          <Heading as="h1" size="xl" textAlign="center">
            Music Match
          </Heading>
          <Text size="s" textAlign="center">
            A tiny application that matches you with your musical soulmate
          </Text>
          
          <form onSubmit={handleSubmit}>
            <VStack spacing={4} align="stretch">
              <FormControl isInvalid={errors.firstName}>
                <FormLabel htmlFor="firstName">First Name</FormLabel>
                <Input
                  id="firstName"
                  name="firstName"
                  value={formData.firstName}
                  onChange={handleChange}
                  variant='flushed'
                  placeholder="Frank"
                />
                <FormErrorMessage>{errors.firstName}</FormErrorMessage>
              </FormControl>
              
              <FormControl isInvalid={errors.lastName}>
                <FormLabel htmlFor="lastName">Last Name</FormLabel>
                <Input
                  id="lastName"
                  name="lastName"
                  value={formData.lastName}
                  onChange={handleChange}
                  variant='flushed'
                  placeholder="Ocean"
                />
                <FormErrorMessage>{errors.lastName}</FormErrorMessage>
              </FormControl>
              
              <FormControl isInvalid={errors.email}>
                <FormLabel htmlFor="email">Email</FormLabel>
                <Input
                  id="email"
                  name="email"
                  type="email"
                  value={formData.email}
                  onChange={handleChange}
                  variant='flushed'
                  placeholder="oddfuture@gmail.com"
                />
                <FormErrorMessage>{errors.email}</FormErrorMessage>
              </FormControl>
              
              <FormControl isInvalid={errors.phoneNumber}>
                <FormLabel htmlFor="phoneNumber">Phone Number</FormLabel>
                <Input
                  id="phoneNumber"
                  name="phoneNumber"
                  type="tel"
                  value={formData.phoneNumber}
                  onChange={handleChange}
                  placeholder="123-456-7890"
                  variant='flushed'
                />
                <FormErrorMessage>{errors.phoneNumber}</FormErrorMessage>
              </FormControl>
              
              <Button
                mt={4}
                colorScheme='spotifygreen'
                // variant='outline'
                type="submit"
                width="full"
                size="lg"
                rightIcon={<MdMusicNote />}
                isDisabled={isLimitReached}
                _hover={isLimitReached ? {} : undefined}
              >
                {isLimitReached ? 'Submissions Closed' : 'Match with Spotify'}
              </Button>
            </VStack>
          </form>
        </VStack>
      </Box>
      </Box>
  );
}

export default App;
