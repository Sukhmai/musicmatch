import { useEffect, useState, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Box,
  Heading,
  Text,
  Spinner,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  VStack,
  Button,
  SimpleGrid,
  Image,
  Badge,
  HStack,
  Link,
  Flex,
} from '@chakra-ui/react';
import { ExternalLinkIcon } from '@chakra-ui/icons';

function Callback() {
  const [status, setStatus] = useState('loading');
  const [error, setError] = useState(null);
  const [result, setResult] = useState(null);
  const navigate = useNavigate();
  const location = useLocation();
  const processedRef = useRef(false);

  useEffect(() => {
    async function processCallback() {
      // Prevent duplicate processing in React.StrictMode
      if (processedRef.current) return;
      processedRef.current = true;
      try {
        // Extract code and state from URL
        const params = new URLSearchParams(location.search);
        const code = params.get('code');
        const state = params.get('state');

        if (!code) {
          throw new Error('No authorization code found in the URL');
        }

        // Exchange code for tokens
        const tokenResponse = await fetch('/api/spotify.v1.SpotifyService/ExchangeToken', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            code: code,
            state: state
          }),
        });

        if (!tokenResponse.ok) {
          const errorData = await tokenResponse.json();
          throw new Error(errorData.message || 'Failed to exchange code for tokens');
        }

        const tokenData = await tokenResponse.json();
        
        // Get user data from localStorage
        const userDataString = localStorage.getItem('userFormData');
        if (!userDataString) {
          throw new Error('User data not found in localStorage');
        }
        
        const userData = JSON.parse(userDataString);
        
        // Save top artists using the access token
        const saveResponse = await fetch('/api/spotify.v1.SpotifyService/SaveTopArtists', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            accessToken: tokenData.accessToken,
            firstName: userData.firstName,
            lastName: userData.lastName,
            email: userData.email,
            number: userData.phoneNumber,
          }),
        });

        if (!saveResponse.ok) {
          const errorData = await saveResponse.json();
          throw new Error(errorData.message || 'Failed to save top artists');
        }

        const saveData = await saveResponse.json();
        console.log(saveData)
        setResult(saveData);
        setStatus('success');
        
        // Clear the localStorage data as it's no longer needed
        localStorage.removeItem('userFormData');
      } catch (err) {
        console.error('Error in callback processing:', err);
        setError(err.message);
        setStatus('error');
      }
    }

    processCallback();
  }, [location, navigate]);

  const handleGoHome = () => {
    navigate('/');
  };

  if (status === 'loading') {
    return (
      <Box bg="#faf0e6" height="100vh" display="flex" justifyContent="center" alignItems="center">
        <VStack spacing={6}>
          <Spinner size="xl" color="green.500" thickness="4px" />
          <Text fontSize="xl">Processing your Spotify data...</Text>
        </VStack>
      </Box>
    );
  }

  if (status === 'error') {
    return (
      <Box bg="#faf0e6" height="100vh" display="flex" justifyContent="center" alignItems="center" p={4}>
        <Alert
          status="error"
          variant="subtle"
          flexDirection="column"
          alignItems="center"
          justifyContent="center"
          textAlign="center"
          height="auto"
          borderRadius="lg"
          p={6}
        >
          <AlertIcon boxSize="40px" mr={0} />
          <AlertTitle mt={4} mb={1} fontSize="lg">
            Something went wrong
          </AlertTitle>
          <AlertDescription maxWidth="sm">
            {error || 'An unexpected error occurred while processing your Spotify data.'}
          </AlertDescription>
          <Button mt={4} colorScheme="green" onClick={handleGoHome}>
            Go Back Home
          </Button>
        </Alert>
      </Box>
    );
  }

  // Artist Card Component
  const ArtistCard = ({ artist }) => {
    // Get the best image (medium size if available)
    const getArtistImage = () => {
      if (!artist.images || artist.images.length === 0) {
        return 'https://via.placeholder.com/300?text=No+Image';
      }
      
      // Sort images by size (medium preferred)
      const sortedImages = [...artist.images].sort((a, b) => {
        // Prefer images around 300px width
        const aDiff = Math.abs(a.width - 300);
        const bDiff = Math.abs(b.width - 300);
        return aDiff - bDiff;
      });
      
      return sortedImages[0].url;
    };

    return (
      <Box 
        borderWidth="1px" 
        borderRadius="lg" 
        overflow="hidden" 
        bg="white" 
        boxShadow="md"
        transition="transform 0.3s"
        _hover={{ transform: 'translateY(-5px)' }}
      >
        <Image 
          src={getArtistImage()} 
          alt={artist.name}
          height="250px"
          width="100%"
          objectFit="cover"
        />
        
        <Box p={4}>
          <Heading as="h3" size="md" mb={2} noOfLines={1}>
            {artist.name}
          </Heading>
          
          {artist.genres && artist.genres.length > 0 && (
            <HStack spacing={2} mb={3} flexWrap="wrap">
              {artist.genres.slice(0, 3).map((genre, index) => (
                <Badge key={index} colorScheme="green" mb={1}>
                  {genre}
                </Badge>
              ))}
              {artist.genres.length > 3 && (
                <Badge colorScheme="gray" mb={1}>+{artist.genres.length - 3}</Badge>
              )}
            </HStack>
          )}
          
          {artist.popularity !== undefined && (
            <Text fontSize="sm" mb={3}>
              Popularity: {artist.popularity}/100
            </Text>
          )}
          
          {artist.spotifyUrl && (
            <Link 
              href={artist.spotifyUrl} 
              isExternal 
              color="green.500"
              fontWeight="bold"
              fontSize="sm"
              display="flex"
              alignItems="center"
            >
              Listen on Spotify <ExternalLinkIcon mx="2px" />
            </Link>
          )}
        </Box>
      </Box>
    );
  };

  return (
    <Box bg="#faf0e6" minHeight="100vh" py={8}>
      <Box
        minWidth="350px"
        maxWidth="1000px" 
        mx="auto" 
        p={8} 
        bg="#fafcff" 
        borderRadius="lg"
        boxShadow="md"
        w="60%"
      >
        <VStack spacing={8} align="stretch">
          <Box textAlign="center">
            <Heading as="h1" size="xl" mb={2}>
              Success!
            </Heading>
            <Text fontSize="lg">
              Your Top Artists have been uploaded to our matching algorithm.
            </Text>
          </Box>
          
          {result && result.uniqueArtists && (
            <Box>
              <Heading as="h2" size="lg" mb={4}>
                Your Top Artists
              </Heading>
              <Text mb={6}>
                You have {result.uniqueArtists.length} unique artists in your top list! Here are some of them:
              </Text>
              
              {result.uniqueArtists.length > 0 && (
                <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} spacing={6}>
                  {result.uniqueArtists.slice(0, 5).map((artist) => (
                    <ArtistCard key={artist.id} artist={artist} />
                  ))}
                </SimpleGrid>
              )}
            </Box>
          )}
          
          <Flex justifyContent="center" mt={6}>
            <Button
              colorScheme="spotifygreen"
              size="lg"
              onClick={handleGoHome}
              width={{ base: "full", md: "auto" }}
              px={8}
            >
              Return Home
            </Button>
          </Flex>
        </VStack>
      </Box>
    </Box>
  );
}

export default Callback;
