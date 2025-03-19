// API utility functions for Spotify Match app

/**
 * Fetch the current user count from the API
 * @returns {Promise<{count: number, maxUsers: number}>}
 */
export const fetchUserCount = async () => {
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
    return {
      count: data.count,
      maxUsers: data.maxUsers
    };
  } catch (error) {
    console.error('Error fetching user count:', error);
    throw error;
  }
};

/**
 * Get the Spotify authentication URL
 * @returns {Promise<{url: string}>}
 */
export const getSpotifyAuthUrl = async () => {
  try {
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
    
    return await response.json();
  } catch (error) {
    console.error('Error getting Spotify auth URL:', error);
    throw error;
  }
};

/**
 * Save user selected artists
 * @param {Object} userData - User data including artist selections
 * @param {string} userData.firstName - User's first name
 * @param {string} userData.lastName - User's last name
 * @param {string} userData.email - User's email
 * @param {string} userData.phoneNumber - User's phone number (optional)
 * @param {string[]} userData.artistIds - Array of selected artist IDs
 * @returns {Promise<Object>} - Response from the API
 */
export const saveUserSelectedArtists = async (userData) => {
  try {
    const response = await fetch('/api/spotify.v1.SpotifyService/SaveUserSelectedArtists', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        firstName: userData.firstName,
        lastName: userData.lastName,
        email: userData.email,
        number: userData.phoneNumber,
        artistIds: userData.artistIds
      }),
    });
    
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || 'Failed to save artist preferences');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Error saving user selected artists:', error);
    throw error;
  }
};

/**
 * Search for artists
 * @param {string} query - Search query
 * @param {number} limit - Maximum number of results to return
 * @param {number} offset - Offset for pagination
 * @returns {Promise<{artists: Array}>} - Array of artist objects
 */
export const searchArtists = async (query, limit = 10, offset = 0) => {
  if (!query.trim()) {
    return { artists: [] };
  }
  
  try {
    const response = await fetch('/api/spotify.v1.SpotifyService/SearchArtists', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        limit,
        offset
      }),
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    return { artists: data.artists || [] };
  } catch (error) {
    console.error('Error searching artists:', error);
    throw error;
  }
};
