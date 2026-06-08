'use client';

import { useEffect } from 'react';
import { MapContainer, TileLayer, CircleMarker, Popup, useMap } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';

const LA_ZIP_COORDS: Record<string, [number, number]> = {
  '90001': [33.9731, -118.2479],
  '90012': [34.0618, -118.2397],
  '90013': [34.0447, -118.2426],
  '90014': [34.0422, -118.2526],
  '90015': [34.0407, -118.2662],
  '90017': [34.0556, -118.2669],
  '90018': [34.028, -118.3172],
  '90019': [34.0481, -118.3348],
  '90028': [34.1016, -118.3268],
  '90034': [34.0294, -118.4115],
};

function FitBounds({ points }: { points: [number, number][] }) {
  const map = useMap();
  useEffect(() => {
    if (points.length > 0) {
      const lats = points.map((p) => p[0]);
      const lngs = points.map((p) => p[1]);
      map.fitBounds([
        [Math.min(...lats) - 0.05, Math.min(...lngs) - 0.05],
        [Math.max(...lats) + 0.05, Math.max(...lngs) + 0.05],
      ]);
    }
  }, [map, points]);
  return null;
}

interface GeoHeatmapProps {
  data: Record<string, number>;
}

export function GeoHeatmap({ data }: GeoHeatmapProps) {
  const entries = Object.entries(data);
  const maxCount = Math.max(...entries.map(([, v]) => v), 1);

  const points = entries
    .map(([zip, count]) => {
      const coords = LA_ZIP_COORDS[zip] ?? [34.0522 + (parseInt(zip.slice(-2)) % 10) * 0.01, -118.2437];
      return { zip, count, coords: coords as [number, number] };
    });

  const allCoords = points.map((p) => p.coords);

  return (
    <div className="h-[500px] overflow-hidden rounded-lg border border-gov-border">
      <MapContainer
        center={[34.0522, -118.2437]}
        zoom={10}
        className="h-full w-full"
        scrollWheelZoom
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <FitBounds points={allCoords} />
        {points.map(({ zip, count, coords }) => {
          const radius = 8 + (count / maxCount) * 24;
          const opacity = 0.3 + (count / maxCount) * 0.5;
          return (
            <CircleMarker
              key={zip}
              center={coords}
              radius={radius}
              pathOptions={{
                color: '#1e3a5f',
                fillColor: '#c9a227',
                fillOpacity: opacity,
                weight: 2,
              }}
            >
              <Popup>
                <strong>ZIP {zip}</strong>
                <br />
                {count} case{count !== 1 ? 's' : ''}
              </Popup>
            </CircleMarker>
          );
        })}
      </MapContainer>
    </div>
  );
}
