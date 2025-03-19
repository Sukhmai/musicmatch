import React from 'react';
import { 
  VStack, 
  FormControl, 
  FormLabel, 
  Button, 
  useToast,
  Divider,
  useDisclosure
} from '@chakra-ui/react';
import UserInfoForm from './UserInfoForm';
import { MdMusicNote, MdSearch } from "react-icons/md";
import { saveUserSelectedArtists } from '../utils/api';
import SelectedArtistsList from './SelectedArtistsList';
import ArtistSearchModal from './ArtistSearchModal';

const ManualArtistForm = ({ 
  formData, 
  errors, 
  handleChange, 
  validateForm, 
  isLimitReached,
  selectedArtists,
  setSelectedArtists,
  fetchUserCount,
  resetForm
}) => {
  const toast = useToast();
  const { isOpen, onOpen, onClose } = useDisclosure();

  // Function to add artist to selected list
  const addArtist = (artist) => {
    // Check if already selected
    if (selectedArtists.some(a => a.id === artist.id)) {
      toast({
        title: 'Artist Already Selected',
        description: `${artist.name} is already in your selected artists.`,
        status: 'info',
        duration: 2000,
        isClosable: true,
      });
      return;
    }
    
    // Add to selected artists (limit to 10)
    if (selectedArtists.length >= 10) {
      toast({
        title: 'Maximum Artists Reached',
        description: 'You can select up to 10 artists. Remove some to add more.',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }
    
    setSelectedArtists(prev => [...prev, artist]);
    
    toast({
      title: 'Artist Added',
      description: `${artist.name} added to your selected artists.`,
      status: 'success',
      duration: 2000,
      isClosable: true,
    });
  };
  
  // Remove artist from selected list
  const removeArtist = (artistId) => {
    setSelectedArtists(prev => prev.filter(artist => artist.id !== artistId));
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
    
    // Validate form
    if (!validateForm()) {
      return;
    }
    
    // Check if artists are selected
    if (selectedArtists.length === 0) {
      toast({
        title: 'No Artists Selected',
        description: 'Please select at least one artist before submitting.',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }
    
    try {
      // Show loading toast
      toast({
        title: 'Submitting',
        description: 'Saving your artist preferences...',
        status: 'info',
        duration: 3000,
        isClosable: true,
      });
      
      // Extract artist IDs
      const artistIds = selectedArtists.map(artist => artist.id);
      
      // Submit to backend
      await saveUserSelectedArtists({
        firstName: formData.firstName,
        lastName: formData.lastName,
        email: formData.email,
        phoneNumber: formData.phoneNumber,
        artistIds: artistIds
      });
      
      // Success message
      toast({
        title: 'Submission Successful',
        description: 'Your artist preferences have been saved. You will be matched soon!',
        status: 'success',
        duration: 5000,
        isClosable: true,
      });
      
      // Update user count after successful submission
      fetchUserCount();
      
      // Reset form state for new submissions
      setSelectedArtists([]);
      
      resetForm();
      
    } catch (error) {
      console.error('Error submitting artist preferences:', error);
      
      toast({
        title: 'Submission Error',
        description: error.message || 'Failed to save your artist preferences. Please try again.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  return (
    <>
      <form onSubmit={handleSubmit}>
        <VStack spacing={4} align="stretch">
          <UserInfoForm 
            formData={formData}
            errors={errors}
            handleChange={handleChange}
            formType="manual"
          />
          
          <Divider my={2} />
          
          <FormControl>
            <FormLabel>Select Your Favorite Artists (Up to 10)</FormLabel>
            <Button 
              leftIcon={<MdSearch />} 
              onClick={onOpen} 
              w="full" 
              colorScheme="gray"
              mb={4}
            >
              Search for Artists
            </Button>
            
            <SelectedArtistsList 
              artists={selectedArtists} 
              onRemoveArtist={removeArtist}
              onReorderArtists={setSelectedArtists}
            />
          </FormControl>
          
          <Button
            mt={4}
            colorScheme='spotifygreen'
            type="submit"
            width="full"
            size="lg"
            rightIcon={<MdMusicNote />}
            isDisabled={isLimitReached || selectedArtists.length === 0}
            _hover={(isLimitReached || selectedArtists.length === 0) ? {} : undefined}
          >
            {isLimitReached ? 'Submissions Closed' : 'Submit Selected Artists'}
          </Button>
        </VStack>
      </form>

      {/* Artist Search Modal */}
      <ArtistSearchModal 
        isOpen={isOpen} 
        onClose={onClose} 
        onSelectArtist={addArtist}
        selectedArtistIds={selectedArtists.map(artist => artist.id)}
      />
    </>
  );
};

export default ManualArtistForm;
