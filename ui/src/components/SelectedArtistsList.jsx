import React from 'react';
import { 
  List, 
  ListItem, 
  HStack, 
  Avatar, 
  Text, 
  Badge, 
  IconButton, 
  Box 
} from '@chakra-ui/react';
import { MdDelete, MdDragHandle } from "react-icons/md";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

// Sortable item component
const SortableItem = ({ artist, index, onRemoveArtist }) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
  } = useSortable({ id: artist.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <ListItem 
      ref={setNodeRef}
      style={style}
      p={2} 
      bg="gray.50" 
      borderRadius="md"
      display="flex"
      alignItems="center"
    >
      <IconButton
        icon={<MdDragHandle />}
        variant="ghost"
        colorScheme="gray"
        aria-label="Reorder"
        cursor="grab"
        size="sm"
        mr={2}
        {...attributes}
        {...listeners}
      />
      <HStack flex="1">
        <Avatar 
          size="sm" 
          src={artist.images && artist.images.length > 0 ? artist.images[0].url : undefined}
          name={artist.name}
        />
        <Text>{artist.name}</Text>
        <Badge ml="auto">{index + 1}</Badge>
      </HStack>
      <IconButton
        icon={<MdDelete />}
        variant="ghost"
        colorScheme="red"
        aria-label="Remove artist"
        onClick={() => onRemoveArtist(artist.id)}
        size="sm"
      />
    </ListItem>
  );
};

const SelectedArtistsList = ({ artists, onRemoveArtist, onReorderArtists }) => {
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleDragEnd = (event) => {
    const { active, over } = event;
    
    if (active.id !== over.id) {
      const oldIndex = artists.findIndex(artist => artist.id === active.id);
      const newIndex = artists.findIndex(artist => artist.id === over.id);
      
      const newOrder = arrayMove(artists, oldIndex, newIndex);
      onReorderArtists(newOrder);
    }
  };

  if (artists.length === 0) {
    return (
      <Box p={4} bg="gray.50" borderRadius="md" textAlign="center">
        <Text color="gray.500">No artists selected yet. Click "Search for Artists" to add some.</Text>
      </Box>
    );
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <SortableContext 
        items={artists.map(artist => artist.id)}
        strategy={verticalListSortingStrategy}
      >
        <List spacing={2} mt={2} border="1px" borderColor="gray.200" borderRadius="md" p={2}>
          {artists.map((artist, index) => (
            <SortableItem 
              key={artist.id}
              artist={artist}
              index={index}
              onRemoveArtist={onRemoveArtist}
            />
          ))}
        </List>
      </SortableContext>
    </DndContext>
  );
};

export default SelectedArtistsList;
