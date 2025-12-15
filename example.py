"""
Example usage of MachineServer API
This script demonstrates how to use the MachineServer API
"""

import requests
import json
import time

# Base URL for the API
BASE_URL = 'http://localhost:5000/api/v1'


def print_response(title, response):
    """Pretty print API response"""
    print(f"\n{'='*60}")
    print(f"{title}")
    print(f"{'='*60}")
    print(f"Status: {response.status_code}")
    print(f"Response: {json.dumps(response.json(), indent=2)}")


def main():
    """Main example workflow"""
    
    print("MachineServer API Example")
    print("=" * 60)
    
    try:
        # 1. Create a simulation
        print("\n1. Creating ARM Cortex-M4 simulation...")
        response = requests.post(
            f'{BASE_URL}/simulation/create',
            json={
                'processor_type': 'arm',
                'config': {
                    'architecture': 'cortex-m4',
                    'frequency': 100000000
                }
            }
        )
        print_response("Create Simulation", response)
        sim_id = response.json().get('simulation_id')
        
        # 2. Start the simulation
        print("\n2. Starting simulation...")
        response = requests.post(f'{BASE_URL}/simulation/{sim_id}/start')
        print_response("Start Simulation", response)
        
        # 3. Load a program
        print("\n3. Loading program...")
        response = requests.post(
            f'{BASE_URL}/execution/load',
            json={
                'simulation_id': sim_id,
                'program_path': '/path/to/program.elf'
            }
        )
        print_response("Load Program", response)
        exec_id = response.json().get('session_id')
        
        # 4. Set a breakpoint
        print("\n4. Setting breakpoint...")
        response = requests.post(
            f'{BASE_URL}/execution/{exec_id}/breakpoint',
            json={'address': '0x08000100'}
        )
        print_response("Set Breakpoint", response)
        
        # 5. Start coverage collection
        print("\n5. Starting coverage collection...")
        response = requests.post(f'{BASE_URL}/coverage/{exec_id}/start')
        print_response("Start Coverage", response)
        cov_id = response.json().get('session_id')
        
        # 6. Execute some steps
        print("\n6. Executing 10 instruction steps...")
        response = requests.post(
            f'{BASE_URL}/execution/{exec_id}/step',
            json={'count': 10}
        )
        print_response("Execute Steps", response)
        
        # 7. Read registers
        print("\n7. Reading registers...")
        response = requests.get(f'{BASE_URL}/execution/{exec_id}/registers')
        print_response("Read Registers", response)
        
        # 8. Read memory
        print("\n8. Reading memory...")
        response = requests.get(
            f'{BASE_URL}/execution/{exec_id}/memory',
            params={'address': '0x08000000', 'size': 64}
        )
        print_response("Read Memory", response)
        
        # 9. Stop coverage and get report
        print("\n9. Stopping coverage and getting report...")
        response = requests.post(f'{BASE_URL}/coverage/{cov_id}/stop')
        print_response("Stop Coverage", response)
        
        response = requests.get(f'{BASE_URL}/coverage/{cov_id}/report')
        print_response("Coverage Report", response)
        
        # 10. Create a co-simulation
        print("\n10. Creating co-simulation...")
        response = requests.post(
            f'{BASE_URL}/cosimulation/create',
            json={
                'components': [
                    {
                        'type': 'processor',
                        'config': {'architecture': 'arm'}
                    },
                    {
                        'type': 'peripheral',
                        'config': {'type': 'uart', 'baudrate': 115200}
                    }
                ]
            }
        )
        print_response("Create Co-Simulation", response)
        cosim_id = response.json().get('session_id')
        
        # 11. Start co-simulation and execute sync steps
        print("\n11. Starting co-simulation...")
        response = requests.post(f'{BASE_URL}/cosimulation/{cosim_id}/start')
        print_response("Start Co-Simulation", response)
        
        print("\n12. Executing synchronized steps...")
        for i in range(3):
            response = requests.post(
                f'{BASE_URL}/cosimulation/{cosim_id}/sync-step',
                json={'time_step_ns': 1000}
            )
            print(f"  Step {i+1}: {response.json()}")
        
        # 12. Get simulation status
        print("\n13. Getting simulation status...")
        response = requests.get(f'{BASE_URL}/simulation/{sim_id}/status')
        print_response("Simulation Status", response)
        
        # 13. Cleanup
        print("\n14. Stopping simulation...")
        response = requests.post(f'{BASE_URL}/simulation/{sim_id}/stop')
        print_response("Stop Simulation", response)
        
        print("\n" + "="*60)
        print("Example completed successfully!")
        print("="*60)
        
    except requests.exceptions.ConnectionError:
        print("\n❌ Error: Could not connect to MachineServer")
        print("Please make sure the server is running with: python app.py")
    except Exception as e:
        print(f"\n❌ Error: {e}")


if __name__ == '__main__':
    main()
