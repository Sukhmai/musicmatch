import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ChakraProvider, extendTheme } from '@chakra-ui/react'
import App from './App'
import Callback from './Callback'

// Create a custom theme with Satoshi font
const theme = extendTheme({
  colors: {
    spotifygreen: {
      50: "#e6f9ef",
      100: "#c3f1d8",
      200: "#9fe9c0",
      300: "#7de0a8",
      400: "#5ad890",
      500: "#5ad890", // Spotify's official green color
      600: "#1DB954",
      700: "#118f3e",
      800: "#0b7a33",
      900: "#056528",
    },
  },
  fonts: {
    heading: "'Satoshi', sans-serif",
    body: "'Satoshi', sans-serif",
  },
})

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ChakraProvider theme={theme}>
      {/* Import Satoshi font */}
      <style jsx global>{`
        @import url('https://api.fontshare.com/v2/css?f=satoshi@400,500,700&display=swap');
      `}</style>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<App />} />
          <Route path="/callback" element={<Callback />} />
        </Routes>
      </BrowserRouter>
    </ChakraProvider>
  </React.StrictMode>
)
