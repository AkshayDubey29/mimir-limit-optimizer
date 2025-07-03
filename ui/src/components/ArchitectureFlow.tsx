import React, { useState, useEffect, useRef } from 'react';

interface FlowData {
  from: string;
  to: string;
  type: string;
  active: boolean;
  throughput: number;
  latency: number;
}

interface ComponentData {
  name: string;
  status: string;
  load: number;
  connections: number;
  endpoint: string;
}

interface ArchitectureFlowProps {
  distributors?: ComponentData[];
  ingesters?: ComponentData[];
  queriers?: ComponentData[];
  query_frontends?: ComponentData[];
  store_gateways?: ComponentData[];
  compactors?: ComponentData[];
  flow?: FlowData[];
}

const ComponentBlock: React.FC<{
  component: ComponentData;
  type: string;
  icon: string;
  color: string;
  position: { x: number; y: number };
  scale: number;
  onDrag?: (newPos: { x: number; y: number }) => void;
  isDragging?: boolean;
}> = ({ component, type, icon, color, position, scale, onDrag, isDragging }) => {
  const [isDraggingLocal, setIsDraggingLocal] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const nodeRef = useRef<HTMLDivElement>(null);

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'running':
      case 'healthy':
        return 'border-green-500 bg-green-50 text-green-800';
      case 'warning':
        return 'border-yellow-500 bg-yellow-50 text-yellow-800';
      case 'critical':
        return 'border-red-500 bg-red-50 text-red-800';
      default:
        return 'border-gray-500 bg-gray-50 text-gray-800';
    }
  };

  const getLoadColor = (load: number) => {
    if (load >= 80) return 'text-red-600 bg-red-100';
    if (load >= 60) return 'text-yellow-600 bg-yellow-100';
    return 'text-green-600 bg-green-100';
  };

  const handleMouseDown = (e: React.MouseEvent) => {
    if (nodeRef.current) {
      const rect = nodeRef.current.getBoundingClientRect();
      setDragOffset({
        x: e.clientX - rect.left - rect.width / 2,
        y: e.clientY - rect.top - rect.height / 2
      });
      setIsDraggingLocal(true);
    }
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (isDraggingLocal && nodeRef.current && onDrag) {
      const container = nodeRef.current.parentElement;
      if (container) {
        const containerRect = container.getBoundingClientRect();
        const newX = ((e.clientX - containerRect.left - dragOffset.x) / containerRect.width) * 100;
        const newY = ((e.clientY - containerRect.top - dragOffset.y) / containerRect.height) * 100;
        onDrag({ 
          x: Math.max(5, Math.min(95, newX)), 
          y: Math.max(5, Math.min(95, newY)) 
        });
      }
    }
  };

  const handleMouseUp = () => {
    setIsDraggingLocal(false);
  };

  useEffect(() => {
    if (isDraggingLocal) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDraggingLocal, dragOffset]);

  return (
    <div 
      ref={nodeRef}
      className={`absolute transform -translate-x-1/2 -translate-y-1/2 transition-all duration-300 hover:scale-110 cursor-move ${
        isDraggingLocal ? 'z-50 scale-110' : 'hover:z-40'
      }`}
      style={{ 
        left: `${position.x}%`, 
        top: `${position.y}%`,
        transform: `translate(-50%, -50%) scale(${scale})`
      }}
      onMouseDown={handleMouseDown}
    >
      <div className={`relative ${color} border-2 rounded-lg p-3 min-w-[140px] shadow-lg hover:shadow-xl transition-all ${getStatusColor(component.status)}`}>
        {/* Component Icon and Name */}
        <div className="text-center mb-2">
          <div className="text-xl mb-1">{icon}</div>
          <div className="font-semibold text-xs">{component.name}</div>
          <div className="text-xs text-gray-600">{type}</div>
        </div>
        
        {/* Status Indicator */}
        <div className="flex items-center justify-center mb-2">
          <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(component.status)}`}>
            {component.status}
          </span>
        </div>
        
        {/* Metrics */}
        <div className="space-y-1 text-xs">
          <div className="flex justify-between">
            <span>Load:</span>
            <span className={`px-1 rounded ${getLoadColor(component.load)}`}>
              {component.load}%
            </span>
          </div>
          <div className="flex justify-between">
            <span>Connections:</span>
            <span className="font-medium">{component.connections}</span>
          </div>
        </div>
        
        {/* Load indicator bar */}
        <div className="mt-2 w-full bg-gray-200 rounded-full h-1">
          <div 
            className={`h-1 rounded-full transition-all duration-500 ${
              component.load >= 80 ? 'bg-red-500' :
              component.load >= 60 ? 'bg-yellow-500' : 'bg-green-500'
            }`}
            style={{ width: `${component.load}%` }}
          ></div>
        </div>
        
        {/* Connection indicator */}
        <div className="absolute -top-1 -right-1">
          <div className={`w-3 h-3 rounded-full ${component.status === 'Running' ? 'bg-green-400 animate-pulse' : 'bg-gray-400'}`}></div>
        </div>
      </div>
    </div>
  );
};

const FlowConnection: React.FC<{
  from: { x: number; y: number };
  to: { x: number; y: number };
  flow: FlowData;
  delay: number;
}> = ({ from, to, flow, delay }) => {
  const [animationPhase, setAnimationPhase] = useState(0);

  useEffect(() => {
    if (!flow.active) return;
    
    const interval = setInterval(() => {
      setAnimationPhase(prev => (prev + 1) % 100);
    }, 50 + delay * 10);

    return () => clearInterval(interval);
  }, [flow.active, delay]);

  return (
    <div className="absolute inset-0 pointer-events-none">
      {/* Connection Line */}
      <svg className="absolute inset-0 w-full h-full">
        <defs>
          <marker
            id={`arrowhead-${delay}`}
            markerWidth="10"
            markerHeight="7"
            refX="9"
            refY="3.5"
            orient="auto"
          >
            <polygon
              points="0 0, 10 3.5, 0 7"
              fill={flow.active ? '#10b981' : '#6b7280'}
              className="transition-colors duration-300"
            />
          </marker>
        </defs>
        
        <line
          x1={`${from.x}%`}
          y1={`${from.y}%`}
          x2={`${to.x}%`}
          y2={`${to.y}%`}
          stroke={flow.active ? '#10b981' : '#6b7280'}
          strokeWidth="2"
          strokeDasharray={flow.active ? "5,5" : "none"}
          strokeDashoffset={flow.active ? -animationPhase : 0}
          markerEnd={`url(#arrowhead-${delay})`}
          className="transition-all duration-300"
        />
      </svg>
      
      {/* Flow Info Tooltip */}
      <div 
        className="absolute transform -translate-x-1/2 -translate-y-1/2"
        style={{ 
          left: `${(from.x + to.x) / 2}%`, 
          top: `${(from.y + to.y) / 2}%` 
        }}
      >
        <div className={`px-2 py-1 rounded text-xs font-medium transition-all duration-300 ${
          flow.active ? 'bg-green-100 text-green-800 shadow-md' : 'bg-gray-100 text-gray-600'
        }`}>
          <div className="font-semibold">{flow.from} → {flow.to}</div>
          <div className="flex space-x-2 text-xs">
            <span>{formatNumber(flow.throughput)}/s</span>
            <span>{flow.latency}ms</span>
          </div>
        </div>
      </div>
    </div>
  );
};

const formatNumber = (num: number) => {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
  return num.toString();
};

const ArchitectureFlow: React.FC<ArchitectureFlowProps> = ({
  distributors = [],
  ingesters = [],
  queriers = [],
  query_frontends = [],
  store_gateways = [],
  compactors = [],
  flow = []
}) => {
  const [selectedComponent, setSelectedComponent] = useState<ComponentData | null>(null);
  const [scale, setScale] = useState(1);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [positions, setPositions] = useState({
    distributors: { x: 20, y: 30 },
    ingesters: { x: 50, y: 20 },
    queriers: { x: 80, y: 30 },
    query_frontends: { x: 80, y: 50 },
    store_gateways: { x: 50, y: 70 },
    compactors: { x: 20, y: 70 }
  });

  const handlePositionChange = (type: string, newPos: { x: number; y: number }) => {
    setPositions(prev => ({
      ...prev,
      [type]: newPos
    }));
  };

  const resetPositions = () => {
    setPositions({
      distributors: { x: 20, y: 30 },
      ingesters: { x: 50, y: 20 },
      queriers: { x: 80, y: 30 },
      query_frontends: { x: 80, y: 50 },
      store_gateways: { x: 50, y: 70 },
      compactors: { x: 20, y: 70 }
    });
  };

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
  };

  const componentTypes = [
    { type: 'distributors', icon: '📥', color: 'bg-blue-50', components: distributors },
    { type: 'ingesters', icon: '💾', color: 'bg-purple-50', components: ingesters },
    { type: 'queriers', icon: '🔍', color: 'bg-green-50', components: queriers },
    { type: 'query_frontends', icon: '🎯', color: 'bg-yellow-50', components: query_frontends },
    { type: 'store_gateways', icon: '🏪', color: 'bg-orange-50', components: store_gateways },
    { type: 'compactors', icon: '🗜️', color: 'bg-red-50', components: compactors }
  ];

  const containerClass = isFullscreen 
    ? "fixed inset-0 z-50 bg-white shadow-2xl"
    : "bg-white rounded-lg shadow-lg border border-gray-200";

  return (
    <div className={containerClass}>
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900 flex items-center">
            🔄 Live Mimir Architecture Flow
            <span className="ml-2 px-2 py-1 bg-green-100 text-green-700 text-xs rounded-full">
              Interactive
            </span>
          </h3>
          <div className="flex items-center space-x-4">
            {/* Zoom Controls */}
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setScale(Math.max(0.5, scale - 0.1))}
                className="p-1 rounded bg-gray-100 hover:bg-gray-200 text-gray-700"
                title="Zoom Out"
              >
                🔍-
              </button>
              <span className="text-sm text-gray-600">{(scale * 100).toFixed(0)}%</span>
              <button
                onClick={() => setScale(Math.min(2, scale + 0.1))}
                className="p-1 rounded bg-gray-100 hover:bg-gray-200 text-gray-700"
                title="Zoom In"
              >
                🔍+
              </button>
            </div>
            
            {/* Control Buttons */}
            <button
              onClick={resetPositions}
              className="px-3 py-1 bg-blue-100 text-blue-700 rounded text-sm hover:bg-blue-200 transition-colors"
            >
              Reset Layout
            </button>
            <button
              onClick={toggleFullscreen}
              className="px-3 py-1 bg-gray-100 text-gray-700 rounded text-sm hover:bg-gray-200 transition-colors"
            >
              {isFullscreen ? '🗗' : '🔲'} {isFullscreen ? 'Exit' : 'Fullscreen'}
            </button>
          </div>
        </div>
        <p className="mt-1 text-sm text-gray-600">
          Interactive data flow visualization - Drag components to reorganize the layout
        </p>
      </div>
      
      <div className="p-6">
        {/* Architecture Diagram */}
        <div className={`relative ${isFullscreen ? 'h-screen' : 'h-[500px]'} bg-gradient-to-br from-gray-50 to-blue-50 rounded-lg border-2 border-dashed border-gray-200 overflow-hidden`}>
          {/* Background Grid */}
          <div className="absolute inset-0 opacity-20">
            <svg className="w-full h-full">
              <pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse">
                <path d="M 20 0 L 0 0 0 20" fill="none" stroke="#e5e7eb" strokeWidth="1"/>
              </pattern>
              <rect width="100%" height="100%" fill="url(#grid)" />
            </svg>
          </div>
          
          {/* Flow Connections */}
          {flow.map((flowItem, index) => {
            const fromType = getComponentTypeFromName(flowItem.from);
            const toType = getComponentTypeFromName(flowItem.to);
            const fromPos = positions[fromType as keyof typeof positions];
            const toPos = positions[toType as keyof typeof positions];
            
            if (fromPos && toPos) {
              return (
                <FlowConnection
                  key={index}
                  from={fromPos}
                  to={toPos}
                  flow={flowItem}
                  delay={index}
                />
              );
            }
            return null;
          })}
          
          {/* Component Blocks */}
          {componentTypes.map(({ type, icon, color, components }) => 
            components.map((component, index) => (
              <ComponentBlock
                key={`${type}-${index}`}
                component={component}
                type={type.replace('_', ' ').toUpperCase()}
                icon={icon}
                color={color}
                position={positions[type as keyof typeof positions]}
                scale={scale}
                onDrag={(newPos) => handlePositionChange(type, newPos)}
              />
            ))
          )}
          
          {/* Legend */}
          <div className="absolute top-2 left-2">
            <div className="bg-white rounded-lg shadow-md p-3 text-xs">
              <div className="font-semibold text-gray-700 mb-2">Legend & Controls</div>
              <div className="space-y-1">
                <div className="flex items-center">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
                  <span>Healthy (Load &lt; 60%)</span>
                </div>
                <div className="flex items-center">
                  <div className="w-2 h-2 bg-yellow-500 rounded-full mr-2"></div>
                  <span>Warning (Load 60-80%)</span>
                </div>
                <div className="flex items-center">
                  <div className="w-2 h-2 bg-red-500 rounded-full mr-2"></div>
                  <span>Critical (Load &gt; 80%)</span>
                </div>
                <div className="border-t pt-2 mt-2">
                  <div className="text-xs text-gray-500">💡 Drag components to move</div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        {/* Flow Statistics */}
        <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-blue-50 rounded-lg p-4">
            <h4 className="font-semibold text-blue-900 mb-2">🔄 Active Flows</h4>
            <div className="text-2xl font-bold text-blue-700">
              {flow.filter(f => f.active).length}
            </div>
            <div className="text-sm text-blue-600">
              of {flow.length} total connections
            </div>
          </div>
          
          <div className="bg-green-50 rounded-lg p-4">
            <h4 className="font-semibold text-green-900 mb-2">📊 Total Throughput</h4>
            <div className="text-2xl font-bold text-green-700">
              {formatNumber(flow.reduce((acc, f) => acc + (f.active ? f.throughput : 0), 0))}/s
            </div>
            <div className="text-sm text-green-600">
              across all active flows
            </div>
          </div>
          
          <div className="bg-purple-50 rounded-lg p-4">
            <h4 className="font-semibold text-purple-900 mb-2">⚡ Avg Latency</h4>
            <div className="text-2xl font-bold text-purple-700">
              {(flow.filter(f => f.active).reduce((acc, f) => acc + f.latency, 0) / Math.max(flow.filter(f => f.active).length, 1)).toFixed(1)}ms
            </div>
            <div className="text-sm text-purple-600">
              end-to-end response time
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// Helper function to map flow names to component types
const getComponentTypeFromName = (name: string): string => {
  const lowerName = name.toLowerCase();
  if (lowerName.includes('distributor') || lowerName.includes('prometheus')) return 'distributors';
  if (lowerName.includes('ingester')) return 'ingesters';
  if (lowerName.includes('querier') && !lowerName.includes('frontend')) return 'queriers';
  if (lowerName.includes('query') && lowerName.includes('frontend')) return 'query_frontends';
  if (lowerName.includes('store') || lowerName.includes('gateway')) return 'store_gateways';
  if (lowerName.includes('compactor')) return 'compactors';
  return 'distributors'; // default
};

export default ArchitectureFlow; 