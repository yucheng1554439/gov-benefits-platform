'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  type Connection,
  type Edge,
  type Node,
  MarkerType,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { api } from '@/lib/api/client';
import type { ApiListResponse, WorkflowTransition } from '@/lib/api/types';
import { formatStatus } from '@/lib/utils';

const STATUS_COLORS: Record<string, string> = {
  submitted: '#3b82f6',
  under_review: '#6366f1',
  need_documents: '#f59e0b',
  eligibility_review: '#8b5cf6',
  supervisor_review: '#ec4899',
  approved: '#166534',
  denied: '#b91c1c',
  appealed: '#b45309',
  closed: '#475569',
};

function transitionsToFlow(transitions: WorkflowTransition[]): { nodes: Node[]; edges: Edge[] } {
  const statuses = new Set<string>();
  transitions.forEach((t) => {
    statuses.add(t.from_status);
    statuses.add(t.to_status);
  });

  const statusList = Array.from(statuses);
  const nodes: Node[] = statusList.map((status, i) => ({
    id: status,
    position: { x: (i % 4) * 220, y: Math.floor(i / 4) * 120 },
    data: { label: formatStatus(status) },
    style: {
      background: STATUS_COLORS[status] ?? '#1e3a5f',
      color: '#fff',
      border: 'none',
      borderRadius: 8,
      padding: '8px 12px',
      fontSize: 12,
      fontWeight: 600,
    },
  }));

  const edges: Edge[] = transitions.map((t, i) => ({
    id: `e-${i}`,
    source: t.from_status,
    target: t.to_status,
    label: t.required_role,
    markerEnd: { type: MarkerType.ArrowClosed },
    style: { stroke: '#475569' },
    labelStyle: { fontSize: 10, fill: '#475569' },
  }));

  return { nodes, edges };
}

export function WorkflowDesigner() {
  const [transitions, setTransitions] = useState<WorkflowTransition[]>([]);
  const [loading, setLoading] = useState(true);

  const initial = useMemo(
    () => transitionsToFlow(transitions),
    [transitions],
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(initial.nodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initial.edges);

  useEffect(() => {
    api
      .get<ApiListResponse<WorkflowTransition>>('/admin/workflow-transitions')
      .then((res) => setTransitions(res.data ?? []))
      .catch(() => setTransitions([]))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    const { nodes: n, edges: e } = transitionsToFlow(transitions);
    setNodes(n);
    setEdges(e);
  }, [transitions, setNodes, setEdges]);

  const onConnect = useCallback(
    (connection: Connection) => setEdges((eds) => addEdge(connection, eds)),
    [setEdges],
  );

  if (loading) {
    return <p className="text-gov-slate">Loading workflow...</p>;
  }

  return (
    <div className="h-[600px] rounded-lg border border-gov-border bg-white">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        fitView
        attributionPosition="bottom-left"
      >
        <Background />
        <Controls />
        <MiniMap />
      </ReactFlow>
    </div>
  );
}
