import * as React from 'react'
import { useState, useRef, useCallback } from 'react'
import { MdSearch } from "react-icons/md";
import SpotifyLogo from '../assets/Spotify_Logo_RGB_Black.png';

// Import Chakra UI components
import {
Text,
Button,
Input,
Image,
VStack,
HStack,
InputGroup,
InputLeftElement,
List,
ListItem,
Avatar,
Spinner,
Badge,
Modal,
ModalOverlay,
ModalContent,
ModalHeader,
ModalBody,
ModalFooter,
ModalCloseButton,
useToast,
Flex,
} from '@chakra-ui/react'

// Artist Search Modal component
const ArtistSearchModal = React.memo(({ isOpen, onClose, onSelectArtist, selectedArtistIds }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [isSearching, setIsSearching] = useState(false);
  const searchTimeoutRef = useRef(null);
  const toast = useToast();
  
  // Search for artists
  const searchArtists = async (query) => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }
    
    setIsSearching(true);
    try {
      const response = await fetch('/api/spotify.v1.SpotifyService/SearchArtists', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: query,
          limit: 10,
          offset: 0
        }),
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data = await response.json();
      setSearchResults(data.artists || []);
    } catch (error) {
      console.error('Error searching artists:', error);
      toast({
        title: 'Search Error',
        description: 'Failed to search for artists. Please try again.',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  };
  
  // Handle search input change with debounce
  const handleSearchChange = (e) => {
    const query = e.target.value;
    setSearchQuery(query);
    
    // Clear any existing timeout
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }
    
    // Set a new timeout for debouncing
    searchTimeoutRef.current = setTimeout(() => {
      searchArtists(query);
    }, 500); // 500ms debounce
  };
  
  // Clean up the search when the modal closes
  const handleClose = useCallback(() => {
    setSearchQuery('');
    setSearchResults([]);
    onClose();
  }, [onClose]);
  
  // SearchInput component
  const SearchInput = () => (
    <InputGroup>
      <InputLeftElement pointerEvents="none">
        <MdSearch color="gray.300" />
      </InputLeftElement>
      <Input
        placeholder="Search for artists..."
        value={searchQuery}
        onChange={handleSearchChange}
        autoFocus
      />
    </InputGroup>
  );
  
  // SearchResults component
  const SearchResults = () => {
    if (isSearching) {
      return (
        <Spinner 
          thickness="4px"
          speed="0.65s"
          emptyColor="gray.200"
          color="green.500"
          size="lg"
        />
      );
    }
    
    if (searchResults.length === 0 && searchQuery) {
      return <Text color="gray.500">No artists found. Try a different search term.</Text>;
    }
    
    return (
      <List spacing={3} w="100%">
        {searchResults.map(artist => (
          <ListItem 
            key={artist.id} 
            p={2} 
            bg="gray.50" 
            borderRadius="md"
            _hover={{ bg: "gray.100" }}
            cursor="pointer"
            onClick={() => onSelectArtist(artist)}
            opacity={selectedArtistIds.includes(artist.id) ? 0.6 : 1}
          >
            <HStack>
              <Avatar 
                size="md" 
                src={artist.images && artist.images.length > 0 ? artist.images[0].url : undefined}
                name={artist.name}
              />
              <VStack align="start" spacing={0}>
                <Text fontWeight="bold">{artist.name}</Text>
                {artist.genres && artist.genres.length > 0 && (
                  <HStack mt={1}>
                    {artist.genres.slice(0, 3).map((genre, idx) => (
                      <Badge key={idx} colorScheme="green" fontSize="xs">{genre}</Badge>
                    ))}
                    {artist.genres.length > 3 && (
                      <Badge colorScheme="gray" fontSize="xs">+{artist.genres.length - 3}</Badge>
                    )}
                  </HStack>
                )}
              </VStack>
              {selectedArtistIds.includes(artist.id) && (
                <Badge colorScheme="green" ml="auto">Already Selected</Badge>
              )}
            </HStack>
          </ListItem>
        ))}
      </List>
    );
  };
  
  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="xl">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>
          <Flex justifyContent="space-between" alignItems="center" w="95%">
            <Text>Search Artists</Text>
            <Image src={SpotifyLogo} alt="Spotify Logo" height="28px" />
          </Flex>
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <VStack spacing={4}>
            <SearchInput />
            <SearchResults />
          </VStack>
        </ModalBody>
        <ModalFooter>
          <Button onClick={handleClose}>Close</Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
});

export default ArtistSearchModal;
