interface ImageCardProps {
  image: string;
  alt?: string;
}

export function ImageCard({ image }: ImageCardProps) {
  return (
    <div className="image-card">
      <button className="remove-image-btn">X</button>
      <div>
        <img src={image} alt="" />
      </div>
    </div>
  );
}
