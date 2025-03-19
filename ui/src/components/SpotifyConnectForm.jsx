import React from 'react';
import { 
  VStack, 
  Button, 
  useToast
} from '@chakra-ui/react';
import UserInfoForm from './UserInfoForm';
import { MdMusicNote } from "react-icons/md";
import { getSpotifyAuthUrl } from '../utils/api';

const SpotifyConnectForm = ({ formData, errors, handleChange, validateForm, isLimitReached }) => {
  const toast = useToast();

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
        
        // Call the backend to get the Spotify auth URL
        const data = await getSpotifyAuthUrl();
        
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
    <form onSubmit={handleSubmit}>
      <VStack spacing={4} align="stretch">
        <UserInfoForm 
          formData={formData}
          errors={errors}
          handleChange={handleChange}
          formType="spotify"
        />
        
        <Button
          mt={4}
          colorScheme='spotifygreen'
          type="submit"
          width="full"
          size="lg"
          rightIcon={<MdMusicNote />}
          isDisabled={isLimitReached}
          _hover={isLimitReached ? {} : undefined}
        >
          {isLimitReached ? 'Submissions Closed' : 'Connect with Spotify'}
        </Button>
      </VStack>
    </form>
  );
};

export default SpotifyConnectForm;
