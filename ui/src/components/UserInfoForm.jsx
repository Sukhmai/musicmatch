import React from 'react';
import { 
  FormControl, 
  FormLabel, 
  Input, 
  FormErrorMessage 
} from '@chakra-ui/react';

const UserInfoForm = ({ formData, errors, handleChange, formType = '' }) => {
  // The formType parameter allows us to add a suffix to input IDs to avoid conflicts
  // when both forms are rendered on the same page
  const idSuffix = formType ? `-${formType}` : '';

  return (
    <>
      <FormControl isInvalid={errors.firstName}>
        <FormLabel htmlFor={`firstName${idSuffix}`}>First Name</FormLabel>
        <Input
          id={`firstName${idSuffix}`}
          name="firstName"
          value={formData.firstName}
          onChange={handleChange}
          variant='flushed'
          placeholder="Frank"
        />
        <FormErrorMessage>{errors.firstName}</FormErrorMessage>
      </FormControl>
      
      <FormControl isInvalid={errors.lastName}>
        <FormLabel htmlFor={`lastName${idSuffix}`}>Last Name</FormLabel>
        <Input
          id={`lastName${idSuffix}`}
          name="lastName"
          value={formData.lastName}
          onChange={handleChange}
          variant='flushed'
          placeholder="Ocean"
        />
        <FormErrorMessage>{errors.lastName}</FormErrorMessage>
      </FormControl>
      
      <FormControl isInvalid={errors.email}>
        <FormLabel htmlFor={`email${idSuffix}`}>Email</FormLabel>
        <Input
          id={`email${idSuffix}`}
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
        <FormLabel htmlFor={`phoneNumber${idSuffix}`}>Phone Number</FormLabel>
        <Input
          id={`phoneNumber${idSuffix}`}
          name="phoneNumber"
          type="tel"
          value={formData.phoneNumber}
          onChange={handleChange}
          placeholder="123-456-7890"
          variant='flushed'
        />
        <FormErrorMessage>{errors.phoneNumber}</FormErrorMessage>
      </FormControl>
    </>
  );
};

export default UserInfoForm;
