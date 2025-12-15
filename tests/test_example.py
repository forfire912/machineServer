"""Example test demonstrating MachineServer API usage"""

import sys
import os

# Add parent directory to path for imports
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))


def test_simulation_flow():
    """Test basic simulation flow"""
    from machineserver.core.simulation_manager import SimulationManager
    
    manager = SimulationManager()
    
    # Create simulation
    result = manager.create_simulation('arm', {'architecture': 'cortex-m4'})
    assert 'simulation_id' in result
    assert result['status'] == 'created'
    
    sim_id = result['simulation_id']
    
    # Start simulation
    result = manager.start_simulation(sim_id)
    assert result['status'] == 'running'
    
    # Get status
    result = manager.get_status(sim_id)
    assert result['status'] == 'running'
    
    # Stop simulation
    result = manager.stop_simulation(sim_id)
    assert result['status'] == 'stopped'
    
    print("✓ Simulation flow test passed")


def test_execution_flow():
    """Test execution and debugging flow"""
    from machineserver.core.execution_engine import ExecutionEngine
    
    engine = ExecutionEngine()
    
    # Load program
    result = engine.load_program('sim_123', '/path/to/program.elf')
    assert 'session_id' in result
    assert result['status'] == 'loaded'
    
    session_id = result['session_id']
    
    # Set breakpoint
    result = engine.set_breakpoint(session_id, '0x08000100')
    assert '0x08000100' in result['breakpoints']
    
    # Execute steps
    result = engine.step(session_id, 5)
    assert result['steps_executed'] == 5
    
    # Read registers
    result = engine.read_registers(session_id)
    assert 'registers' in result
    
    # Read memory
    result = engine.read_memory(session_id, '0x08000000', 256)
    assert result['size'] == 256
    
    print("✓ Execution flow test passed")


def test_coverage_flow():
    """Test coverage analysis flow"""
    from machineserver.core.coverage_analyzer import CoverageAnalyzer
    
    analyzer = CoverageAnalyzer()
    
    # Start coverage
    result = analyzer.start_coverage('exec_123')
    assert 'session_id' in result
    assert result['status'] == 'collecting'
    
    session_id = result['session_id']
    
    # Stop coverage
    result = analyzer.stop_coverage(session_id)
    assert result['status'] == 'completed'
    assert 'coverage_percentage' in result
    
    # Get report
    result = analyzer.get_report(session_id)
    assert 'summary' in result
    assert 'file_coverage' in result
    
    # Export coverage
    result = analyzer.export_coverage(session_id, 'json')
    assert result['status'] == 'exported'
    
    print("✓ Coverage flow test passed")


def test_cosimulation_flow():
    """Test co-simulation flow"""
    from machineserver.core.cosim_coordinator import CoSimCoordinator
    
    coordinator = CoSimCoordinator()
    
    # Create co-simulation
    components = [
        {'type': 'processor', 'config': {'arch': 'arm'}},
        {'type': 'peripheral', 'config': {'type': 'uart'}}
    ]
    result = coordinator.create_cosimulation(components)
    assert 'session_id' in result
    assert len(result['components']) == 2
    
    session_id = result['session_id']
    
    # Start co-simulation
    result = coordinator.start_cosimulation(session_id)
    assert result['status'] == 'running'
    
    # Sync step
    result = coordinator.sync_step(session_id, 1000)
    assert result['sync_count'] == 1
    assert result['time_ns'] == 1000
    
    # Stop co-simulation
    result = coordinator.stop_cosimulation(session_id)
    assert result['status'] == 'stopped'
    
    print("✓ Co-simulation flow test passed")


if __name__ == '__main__':
    print("Running MachineServer tests...\n")
    
    test_simulation_flow()
    test_execution_flow()
    test_coverage_flow()
    test_cosimulation_flow()
    
    print("\n✓ All tests passed!")
