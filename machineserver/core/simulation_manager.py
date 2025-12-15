"""Simulation Manager - Core simulation management module"""

import uuid
from typing import Dict, Any, Optional
from datetime import datetime
from ..utils.logger import logger


class SimulationInstance:
    """Represents a single simulation instance"""
    
    def __init__(self, simulation_id: str, processor_type: str, config: Dict[str, Any]):
        self.id = simulation_id
        self.processor_type = processor_type
        self.config = config
        self.status = 'created'
        self.created_at = datetime.now()
        self.started_at = None
        self.stopped_at = None
        self.cycles = 0
        self.time_ns = 0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert instance to dictionary"""
        return {
            'id': self.id,
            'processor_type': self.processor_type,
            'config': self.config,
            'status': self.status,
            'created_at': self.created_at.isoformat(),
            'started_at': self.started_at.isoformat() if self.started_at else None,
            'stopped_at': self.stopped_at.isoformat() if self.stopped_at else None,
            'cycles': self.cycles,
            'time_ns': self.time_ns
        }


class SimulationManager:
    """Manages embedded processor simulation instances"""
    
    def __init__(self):
        self.simulations: Dict[str, SimulationInstance] = {}
        logger.info("SimulationManager initialized")
    
    def create_simulation(
        self,
        processor_type: str,
        config: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Create a new simulation instance
        
        Args:
            processor_type: Type of processor (e.g., 'arm', 'riscv', 'x86')
            config: Simulation configuration
            
        Returns:
            Dictionary containing simulation_id and status
        """
        simulation_id = f"sim_{uuid.uuid4().hex[:8]}"
        
        instance = SimulationInstance(simulation_id, processor_type, config)
        self.simulations[simulation_id] = instance
        
        logger.info(f"Created simulation {simulation_id} for {processor_type}")
        
        return {
            'simulation_id': simulation_id,
            'status': 'created',
            'processor_type': processor_type
        }
    
    def start_simulation(self, simulation_id: str) -> Dict[str, Any]:
        """Start a simulation instance
        
        Args:
            simulation_id: ID of simulation to start
            
        Returns:
            Status dictionary
        """
        if simulation_id not in self.simulations:
            return {'error': 'Simulation not found', 'status': 'error'}
        
        instance = self.simulations[simulation_id]
        instance.status = 'running'
        instance.started_at = datetime.now()
        
        logger.info(f"Started simulation {simulation_id}")
        
        return {
            'simulation_id': simulation_id,
            'status': 'running',
            'started_at': instance.started_at.isoformat()
        }
    
    def stop_simulation(self, simulation_id: str) -> Dict[str, Any]:
        """Stop a simulation instance
        
        Args:
            simulation_id: ID of simulation to stop
            
        Returns:
            Status dictionary
        """
        if simulation_id not in self.simulations:
            return {'error': 'Simulation not found', 'status': 'error'}
        
        instance = self.simulations[simulation_id]
        instance.status = 'stopped'
        instance.stopped_at = datetime.now()
        
        logger.info(f"Stopped simulation {simulation_id}")
        
        return {
            'simulation_id': simulation_id,
            'status': 'stopped',
            'stopped_at': instance.stopped_at.isoformat()
        }
    
    def get_status(self, simulation_id: str) -> Dict[str, Any]:
        """Get simulation status
        
        Args:
            simulation_id: ID of simulation
            
        Returns:
            Status dictionary
        """
        if simulation_id not in self.simulations:
            return {'error': 'Simulation not found', 'status': 'error'}
        
        instance = self.simulations[simulation_id]
        return instance.to_dict()
    
    def list_simulations(self) -> Dict[str, Any]:
        """List all simulation instances
        
        Returns:
            Dictionary containing list of simulations
        """
        return {
            'simulations': [
                instance.to_dict()
                for instance in self.simulations.values()
            ],
            'count': len(self.simulations)
        }
    
    def delete_simulation(self, simulation_id: str) -> Dict[str, Any]:
        """Delete a simulation instance
        
        Args:
            simulation_id: ID of simulation to delete
            
        Returns:
            Status dictionary
        """
        if simulation_id not in self.simulations:
            return {'error': 'Simulation not found', 'status': 'error'}
        
        del self.simulations[simulation_id]
        logger.info(f"Deleted simulation {simulation_id}")
        
        return {
            'simulation_id': simulation_id,
            'status': 'deleted'
        }
