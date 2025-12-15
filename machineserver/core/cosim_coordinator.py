"""Co-Simulation Coordinator - System-level collaborative simulation module"""

import uuid
from typing import Dict, Any, List
from datetime import datetime
from ..utils.logger import logger


class CoSimComponent:
    """Represents a component in co-simulation"""
    
    def __init__(self, component_id: str, component_type: str, config: Dict[str, Any]):
        self.id = component_id
        self.type = component_type
        self.config = config
        self.status = 'initialized'
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert component to dictionary"""
        return {
            'id': self.id,
            'type': self.type,
            'config': self.config,
            'status': self.status
        }


class CoSimSession:
    """Represents a co-simulation session"""
    
    def __init__(self, session_id: str):
        self.id = session_id
        self.components: List[CoSimComponent] = []
        self.status = 'created'
        self.created_at = datetime.now()
        self.started_at = None
        self.sync_count = 0
        self.time_ns = 0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert session to dictionary"""
        return {
            'id': self.id,
            'status': self.status,
            'created_at': self.created_at.isoformat(),
            'started_at': self.started_at.isoformat() if self.started_at else None,
            'components': [comp.to_dict() for comp in self.components],
            'sync_count': self.sync_count,
            'time_ns': self.time_ns
        }


class CoSimCoordinator:
    """Manages system-level collaborative simulation"""
    
    def __init__(self):
        self.sessions: Dict[str, CoSimSession] = {}
        logger.info("CoSimCoordinator initialized")
    
    def create_cosimulation(
        self,
        components: List[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """Create a co-simulation session with multiple components
        
        Args:
            components: List of component configurations
            
        Returns:
            Co-simulation session information
        """
        session_id = f"cosim_{uuid.uuid4().hex[:8]}"
        
        session = CoSimSession(session_id)
        
        # Create components
        for comp_config in components:
            comp_id = f"comp_{uuid.uuid4().hex[:8]}"
            component = CoSimComponent(
                comp_id,
                comp_config.get('type', 'unknown'),
                comp_config.get('config', {})
            )
            session.components.append(component)
        
        self.sessions[session_id] = session
        
        logger.info(f"Created co-simulation session {session_id} with {len(components)} components")
        
        return {
            'session_id': session_id,
            'status': 'created',
            'components': [comp.to_dict() for comp in session.components]
        }
    
    def start_cosimulation(self, session_id: str) -> Dict[str, Any]:
        """Start co-simulation
        
        Args:
            session_id: ID of co-simulation session
            
        Returns:
            Status dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Co-simulation session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        session.status = 'running'
        session.started_at = datetime.now()
        
        # Update component status
        for component in session.components:
            component.status = 'running'
        
        logger.info(f"Started co-simulation session {session_id}")
        
        return {
            'session_id': session_id,
            'status': 'running',
            'started_at': session.started_at.isoformat()
        }
    
    def sync_step(
        self,
        session_id: str,
        time_step_ns: int = 1000
    ) -> Dict[str, Any]:
        """Execute synchronized step across all components
        
        Args:
            session_id: ID of co-simulation session
            time_step_ns: Time step in nanoseconds
            
        Returns:
            Synchronization result
        """
        if session_id not in self.sessions:
            return {'error': 'Co-simulation session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        session.sync_count += 1
        session.time_ns += time_step_ns
        
        logger.info(f"Executed sync step {session.sync_count} in session {session_id}")
        
        return {
            'session_id': session_id,
            'sync_count': session.sync_count,
            'time_ns': session.time_ns,
            'time_step_ns': time_step_ns,
            'status': 'synchronized'
        }
    
    def stop_cosimulation(self, session_id: str) -> Dict[str, Any]:
        """Stop co-simulation
        
        Args:
            session_id: ID of co-simulation session
            
        Returns:
            Status dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Co-simulation session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        session.status = 'stopped'
        
        # Update component status
        for component in session.components:
            component.status = 'stopped'
        
        logger.info(f"Stopped co-simulation session {session_id}")
        
        return {
            'session_id': session_id,
            'status': 'stopped'
        }
    
    def get_status(self, session_id: str) -> Dict[str, Any]:
        """Get co-simulation status
        
        Args:
            session_id: ID of co-simulation session
            
        Returns:
            Session status
        """
        if session_id not in self.sessions:
            return {'error': 'Co-simulation session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        return session.to_dict()
    
    def exchange_data(
        self,
        session_id: str,
        source_component: str,
        target_component: str,
        data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Exchange data between components
        
        Args:
            session_id: ID of co-simulation session
            source_component: Source component ID
            target_component: Target component ID
            data: Data to exchange
            
        Returns:
            Exchange result
        """
        if session_id not in self.sessions:
            return {'error': 'Co-simulation session not found', 'status': 'error'}
        
        logger.info(f"Data exchange from {source_component} to {target_component}")
        
        return {
            'session_id': session_id,
            'source': source_component,
            'target': target_component,
            'status': 'transferred'
        }
