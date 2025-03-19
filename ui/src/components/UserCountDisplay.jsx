import React from 'react';
import { Text, Box } from '@chakra-ui/react';

const UserCountDisplay = ({ userCount, maxUsers, isLimitReached, isAnimating }) => {
  return (
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
  );
};

export default UserCountDisplay;
