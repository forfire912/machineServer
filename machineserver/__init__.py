"""
MachineServer - 统一仿真微服务平台
A unified simulation microservice platform for embedded systems
"""

__version__ = "0.1.0"
__author__ = "MachineServer Team"

from .core.simulation_manager import SimulationManager
from .core.execution_engine import ExecutionEngine
from .core.coverage_analyzer import CoverageAnalyzer
from .core.cosim_coordinator import CoSimCoordinator

__all__ = [
    "SimulationManager",
    "ExecutionEngine",
    "CoverageAnalyzer",
    "CoSimCoordinator",
]
