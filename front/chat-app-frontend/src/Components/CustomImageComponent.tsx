interface ImageCardProps {
  image: string;
  placeholder?: string;
  handleRemove: (fileUrl: string) => void;
}

export function ImageCard({
  image,
  placeholder = "",
  handleRemove,
}: ImageCardProps) {
  return (
    <div className="image-card">
      <button className="remove-image-btn" onClick={() => handleRemove(image)}>
        X
      </button>
      <div>
        <img src={image} alt={placeholder} />
      </div>
    </div>
  );
}
