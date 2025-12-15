"""Execution Engine - Program execution and debugging module"""

import uuid
from typing import Dict, Any, List, Optional
from ..utils.logger import logger


class ExecutionSession:
    """Represents a program execution session"""
    
    def __init__(self, session_id: str, simulation_id: str, program_path: str):
        self.id = session_id
        self.simulation_id = simulation_id
        self.program_path = program_path
        self.status = 'loaded'
        self.breakpoints: List[str] = []
        self.pc = "0x00000000"  # Program counter
        self.registers: Dict[str, str] = {}
        self.step_count = 0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert session to dictionary"""
        return {
            'id': self.id,
            'simulation_id': self.simulation_id,
            'program_path': self.program_path,
            'status': self.status,
            'breakpoints': self.breakpoints,
            'pc': self.pc,
            'step_count': self.step_count
        }


class ExecutionEngine:
    """Manages program execution and debugging"""
    
    def __init__(self):
        self.sessions: Dict[str, ExecutionSession] = {}
        logger.info("ExecutionEngine initialized")
    
    def load_program(
        self,
        simulation_id: str,
        program_path: str
    ) -> Dict[str, Any]:
        """Load a program for execution
        
        Args:
            simulation_id: ID of simulation instance
            program_path: Path to program binary
            
        Returns:
            Session information dictionary
        """
        session_id = f"exec_{uuid.uuid4().hex[:8]}"
        
        session = ExecutionSession(session_id, simulation_id, program_path)
        self.sessions[session_id] = session
        
        logger.info(f"Loaded program {program_path} in session {session_id}")
        
        return {
            'session_id': session_id,
            'simulation_id': simulation_id,
            'program_path': program_path,
            'status': 'loaded'
        }
    
    def step(self, session_id: str, count: int = 1) -> Dict[str, Any]:
        """Execute instruction steps
        
        Args:
            session_id: ID of execution session
            count: Number of steps to execute
            
        Returns:
            Execution status dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        session.step_count += count
        session.status = 'paused'
        
        logger.info(f"Executed {count} steps in session {session_id}")
        
        return {
            'session_id': session_id,
            'steps_executed': count,
            'total_steps': session.step_count,
            'pc': session.pc,
            'status': 'paused'
        }
    
    def run(self, session_id: str) -> Dict[str, Any]:
        """Run program until breakpoint or completion
        
        Args:
            session_id: ID of execution session
            
        Returns:
            Execution status dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        session.status = 'running'
        
        logger.info(f"Running program in session {session_id}")
        
        return {
            'session_id': session_id,
            'status': 'running'
        }
    
    def set_breakpoint(self, session_id: str, address: str) -> Dict[str, Any]:
        """Set a breakpoint at specified address
        
        Args:
            session_id: ID of execution session
            address: Memory address for breakpoint
            
        Returns:
            Breakpoint information dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        
        if address not in session.breakpoints:
            session.breakpoints.append(address)
            logger.info(f"Set breakpoint at {address} in session {session_id}")
        
        return {
            'session_id': session_id,
            'address': address,
            'breakpoints': session.breakpoints
        }
    
    def remove_breakpoint(self, session_id: str, address: str) -> Dict[str, Any]:
        """Remove a breakpoint
        
        Args:
            session_id: ID of execution session
            address: Memory address of breakpoint
            
        Returns:
            Status dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        
        if address in session.breakpoints:
            session.breakpoints.remove(address)
            logger.info(f"Removed breakpoint at {address} in session {session_id}")
        
        return {
            'session_id': session_id,
            'address': address,
            'breakpoints': session.breakpoints
        }
    
    def read_registers(self, session_id: str) -> Dict[str, Any]:
        """Read processor registers
        
        Args:
            session_id: ID of execution session
            
        Returns:
            Register values dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        
        # Mock register values
        registers = {
            'r0': '0x00000000',
            'r1': '0x00000001',
            'r2': '0x00000002',
            'r3': '0x00000003',
            'pc': session.pc,
            'sp': '0x20000800',
            'lr': '0x08000100'
        }
        
        return {
            'session_id': session_id,
            'registers': registers
        }
    
    def read_memory(
        self,
        session_id: str,
        address: str,
        size: int
    ) -> Dict[str, Any]:
        """Read memory contents
        
        Args:
            session_id: ID of execution session
            address: Starting memory address
            size: Number of bytes to read
            
        Returns:
            Memory contents dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Session not found', 'status': 'error'}
        
        # Mock memory contents
        memory_data = '00' * size
        
        return {
            'session_id': session_id,
            'address': address,
            'size': size,
            'data': memory_data
        }
    
    def get_status(self, session_id: str) -> Dict[str, Any]:
        """Get execution session status
        
        Args:
            session_id: ID of execution session
            
        Returns:
            Session status dictionary
        """
        if session_id not in self.sessions:
            return {'error': 'Session not found', 'status': 'error'}
        
        session = self.sessions[session_id]
        return session.to_dict()
