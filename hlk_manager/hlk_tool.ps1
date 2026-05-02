[CmdletBinding()]
param([string]$cmdline)

function New-Task($name, $stage, $status, $taskerrormessage, $tasktype, $childtasks)
{
    $task = New-Object PSObject
    $task | Add-Member -type NoteProperty -Name name -Value $name
    $task | Add-Member -type NoteProperty -Name stage -Value $stage
    $task | Add-Member -type NoteProperty -Name status -Value $status
    $task | Add-Member -type NoteProperty -Name taskerrormessage -Value $taskerrormessage
    $task | Add-Member -type NoteProperty -Name tasktype -Value $tasktype
    $task | Add-Member -type NoteProperty -Name childtasks -Value $childtasks
    return $task
}

function New-PackageProgressInfo($current, $maximum, $message)
{
    $packageprogressinfo = New-Object PSObject
    $packageprogressinfo | Add-Member -type NoteProperty -Name current -Value $current
    $packageprogressinfo | Add-Member -type NoteProperty -Name maximum -Value $maximum
    $packageprogressinfo | Add-Member -type NoteProperty -Name message -Value $message
    return $packageprogressinfo
}

function New-ProjectPackage($name, $projectpackagepath)
{
    $projectpackage = New-Object PSObject
    $projectpackage | Add-Member -type NoteProperty -Name name -Value $name
    $projectpackage | Add-Member -type NoteProperty -Name projectpackagepath -Value $projectpackagepath
    return $projectpackage
}

function New-TestResultLogsZip($testname, $testid, $status, $logszippath)
{
    $testresultlogszip = New-Object PSObject
    $testresultlogszip | Add-Member -type NoteProperty -Name testname -Value $testname
    $testresultlogszip | Add-Member -type NoteProperty -Name testid -Value $testid
    $testresultlogszip | Add-Member -type NoteProperty -Name status -Value $status
    $testresultlogszip | Add-Member -type NoteProperty -Name logszippath -Value $logszippath
    return $testresultlogszip
}

function New-TestResult($name, $completiontime, $scheduletime, $starttime, $status, $arefiltersapplied, $target, $tasks)
{
    $testresult = New-Object PSObject
    $testresult | Add-Member -type NoteProperty -Name name -Value $name
    $testresult | Add-Member -type NoteProperty -Name completiontime -Value $completiontime
    $testresult | Add-Member -type NoteProperty -Name scheduletime -Value $scheduletime
    $testresult | Add-Member -type NoteProperty -Name starttime -Value $starttime
    $testresult | Add-Member -type NoteProperty -Name status -Value $status
    $testresult | Add-Member -type NoteProperty -Name arefiltersapplied -Value $arefiltersapplied
    $testresult | Add-Member -type NoteProperty -Name target -Value $target
    $testresult | Add-Member -type NoteProperty -Name tasks -Value $tasks
    return $testresult
}

function New-FilterResult($appliedfilterson)
{
    $filterresult = New-Object PSObject
    $filterresult | Add-Member -type NoteProperty -Name appliedfilterson -Value $appliedfilterson
    return $filterresult
}

function New-Test($name, $id, $testtype, $estimatedruntime, $requiresspecialconfiguration, $requiressupplementalcontent, $scheduleoptions, $status, $executionstate)
{
    $test = New-Object PSObject
    $test | Add-Member -type NoteProperty -Name name -Value $name
    $test | Add-Member -type NoteProperty -Name id -Value $id
    $test | Add-Member -type NoteProperty -Name testtype -Value $testtype
    $test | Add-Member -type NoteProperty -Name estimatedruntime -Value $estimatedruntime
    $test | Add-Member -type NoteProperty -Name requiresspecialconfiguration -Value $requiresspecialconfiguration
    $test | Add-Member -type NoteProperty -Name requiressupplementalcontent -Value $requiressupplementalcontent
    $test | Add-Member -type NoteProperty -Name scheduleoptions -Value $scheduleoptions
    $test | Add-Member -type NoteProperty -Name status -Value $status
    $test | Add-Member -type NoteProperty -Name executionstate -Value $executionstate
    return $test
}

function New-ProductInstanceTarget($name, $key, $machine)
{
    $productinstancetarget = New-Object PSObject
    $productinstancetarget | Add-Member -type NoteProperty -Name name -Value $name
    $productinstancetarget | Add-Member -type NoteProperty -Name key -Value $key
    $productinstancetarget | Add-Member -type NoteProperty -Name machine -Value $machine
    return $productinstancetarget
}

function New-ProductInstance($name, $osplatform, $targetedpool, $targets)
{
    $productinstance = New-Object PSObject
    $productinstance | Add-Member -type NoteProperty -Name name -Value $name
    $productinstance | Add-Member -type NoteProperty -Name osplatform -Value $osplatform
    $productinstance | Add-Member -type NoteProperty -Name targetedpool -Value $targetedpool
    $productinstance | Add-Member -type NoteProperty -Name targets -Value $targets
    return $productinstance
}

function New-Project($name, $creationtime, $modifiedtime, $status, $productinstances)
{
    $project = New-Object PSObject
    $project | Add-Member -type NoteProperty -Name name -Value $name
    $project | Add-Member -type NoteProperty -Name creationtime -Value $creationtime
    $project | Add-Member -type NoteProperty -Name modifiedtime -Value $modifiedtime
    $project | Add-Member -type NoteProperty -Name status -Value $status
    $project | Add-Member -type NoteProperty -Name productinstances -Value $productinstances
    return $project
}

function New-Target($name, $key, $type)
{
    $target = New-Object PSObject
    $target | Add-Member -type NoteProperty -Name name -Value $name
    $target | Add-Member -type NoteProperty -Name key -Value $key
    $target | Add-Member -type NoteProperty -Name type -value $type
    return $target
}

function New-Machine($name, $state, $lastheartbeat)
{
    $machine = New-Object PSObject
    $machine | Add-Member -type NoteProperty -Name name -Value $name
    $machine | Add-Member -type NoteProperty -Name state -Value $state
    $machine | Add-Member -type NoteProperty -Name lastheartbeat -Value $lastheartbeat
    return $machine
}

function New-Pool($name, $machines)
{
    $pool = New-Object PSObject
    $pool | Add-Member -type NoteProperty -Name name -Value $name
    $pool | Add-Member -type NoteProperty -Name machines -Value $machines
    return $pool
}

function New-ActionResult($content, $exception = $nil)
{
    $actionresult = New-Object PSObject
    if ( [String]::IsNullOrEmpty($exception))
    {
        $actionresult | Add-Member -type NoteProperty -Name result -Value $true
        if (-Not [String]::IsNullOrEmpty($content))
        {
            $jsoncontent = (ConvertFrom-Json $content)
            if ($jsoncontent -is [System.Object[]])
            {
                $actionresult | Add-Member -type NoteProperty -Name content -Value $jsoncontent.SyncRoot
            }
            else
            {
                $actionresult | Add-Member -type NoteProperty -Name content -Value $jsoncontent
            }
        }
    }
    else
    {
        $actionresult | Add-Member -type NoteProperty -Name result -Value $false
        if ( [String]::IsNullOrEmpty($exception.InnerException))
        {
            $actionresult | Add-Member -type NoteProperty -Name message -Value $exception.Message
        }
        else
        {
            $actionresult | Add-Member -type NoteProperty -Name message -Value $exception.InnerException.Message
        }
    }
    return $actionresult
}

function Write-Tasks-Log-File($tasks, $parentPath)
{
    for ($i = 0; $i -lt $tasks.Count; $i++) {
        $task = $tasks[$i]
        $taskName = ($i + 1).ToString() + ". " + $task.Status.ToString()
        $taskName = (Check-File-Name $taskName)
        $ppath = [System.IO.Path]::Combine($parentPath, $taskName).ToString()
        New-Item -ItemType Directory -Force -Path $ppath | Out-Null
        $taskNameFile = ([System.IO.Path]::Combine($ppath, "task_name_and_errmsg.txt").ToString())
        Set-Content -Path $taskNameFile -Value ($task.Name.ToString())
        if (-Not ([String]::IsNullOrEmpty($task.TaskErrorMessage)))
        {
            Add-Content -Path $taskNameFile -Value $task.TaskErrorMessage
        }
        foreach ($log in $task.GetLogFiles())
        {
            $logName = Check-File-Name $log.Name
            $path = [System.IO.Path]::Combine($ppath, $logName).ToString()
            $log.WriteLogTo($path)
        }
        $childTasks = $task.GetChildTasks()
        Write-Tasks-Log-File $childTasks $ppath
    }
}

function Check-File-Name($name)
{
    $name = $name -replace '\\', ''
    $name = $name -replace '/', ''
    $name = $name -replace ':', ''
    $name = $name -replace '\*', ''
    $name = $name -replace '\?', ''
    $name = $name -replace '"', ''
    $name = $name -replace '<', ''
    $name = $name -replace '>', ''
    $name = $name -replace '\|', ''
    return $name
}

function Fail-Host($msg)
{
    $actionresult = New-Object PSObject
    $actionresult | Add-Member -type NoteProperty -Name result -Value $false
    $actionresult | Add-Member -type NoteProperty -Name message -Value $msg
    $output = $actionresult | ConvertTo-Json -Depth $MaxJsonDepth -Compress
    Write-Output $output
    exit 0
}

# 列出所有测试池。
function List-Pools
{
    $poolslist = New-Object System.Collections.ArrayList
    foreach ($Pool in $RootPool.GetChildPools())
    {
        $machineslist = New-Object System.Collections.ArrayList
        $Machines = $Pool.GetMachines()
        foreach ($Machine in $Machines)
        {
            $machineslist.Add((New-Machine $Machine.Name $Machine.Status.ToString() $Machine.LastHeartBeat.ToString())) | Out-Null
        }
        $poolslist.Add((New-Pool $Pool.Name $machineslist)) | Out-Null
    }
    ConvertTo-Json @($poolslist) -Depth $MaxJsonDepth -Compress
}

# 创建测试池。
function Create-Pool
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$pool)
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    ConvertTo-JSON -Depth $MaxJsonDepth  $RootPool.CreateChildPool($pool) | Out-Null
}

# 删除测试池。
function Delete-Pool
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$pool)
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Provided pool's name is not valid, aborting..."
    }
    ConvertTo-JSON -Depth $MaxJsonDepth  $RootPool.DeleteChildPool($WntdPool)
}

# 移动机器到别的池子。
function Move-Machine
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$machine, [Parameter(Position = 2)][String]$from, [Parameter(Position = 3)][String]$to)
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($from))
    {
        throw "Please provide a source pool's name."
    }
    if ( [String]::IsNullOrEmpty($to))
    {
        throw "Please provide a destination pool's name."
    }
    if (-Not ($WntdFromPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $from }))
    {
        throw "Provided source pool's name is not valid, aborting..."
    }
    if (-Not ($WntdToPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $to }))
    {
        throw "Provided destination pool's name is not valid, aborting..."
    }
    if (-Not ($WntdMachine = $WntdFromPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "Provided machines's name is not valid, aborting..."
    }
    ConvertTo-JSON -Depth $MaxJsonDepth ($WntdFromPool.MoveMachineTo($WntdMachine, $WntdToPool))
}

# 设置机器状态。
function Set-Machine-State
{
    [CmdletBinding()]
    param([Int]$timeout = -1, [Parameter(Position = 1)][String]$machine, [Parameter(Position = 2)][String]$pool, [Parameter(Position = 3)][String]$state)
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if ( [String]::IsNullOrEmpty($state))
    {
        throw "Please provide a state."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Provided pool's name is not valid, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "Provided machines's name is not valid, aborting..."
    }
    if (-Not ($timeout -eq -1))
    {
        $timeout = $timeout * 1000
    }
    switch ($state)
    {
        "Ready" {
            if (-Not $WntdMachine.SetMachineStatus([Microsoft.Windows.Kits.Hardware.ObjectModel.MachineStatus]::Ready, $timeout))
            {
                throw "Unable to change machine state, timed out."
            }
        }
        "NotReady" {
            if (-Not $WntdMachine.SetMachineStatus([Microsoft.Windows.Kits.Hardware.ObjectModel.MachineStatus]::NotReady, $timeout))
            {
                throw "Unable to change machine state, timed out."
            }
        }
        default {
            throw "Provided desired machines's sate is not valid, aborting..."
        }
    }

    return "{}"
}

# 删除测试机。
function Delete-Machine
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$machine, [Parameter(Position = 2)][String]$pool)
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Provided pool's name is not valid, aborting..."
    }
    ConvertTo-JSON -Depth $MaxJsonDepth  $WntdPool.DeleteMachine($machine)
}

# 列出测试机器的测试目标。
function List-Machine-Targets
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$machine, [Parameter(Position = 2)][String]$pool)
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "please provide a machine's name"
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "please provide a pool's name"
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "did not find pool $pool in Root pool"
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "the test machine was not found"
    }
    $targetslist = New-Object System.Collections.ArrayList
    foreach ($TestTarget in $WntdMachine.GetTestTargets())
    {
        $targetslist.Add((New-Target $TestTarget.Name $TestTarget.Key $TestTarget.TargetType)) | Out-Null
    }
    return (ConvertTo-Json @($targetslist) -Depth $MaxJsonDepth -Compress)
}

# 列出所有测试项目。
function List-Projects
{
    $projectslist = New-Object System.Collections.ArrayList
    foreach ($ProjectName in $Manager.GetProjectNames())
    {
        $Project = $Manager.GetProject($ProjectName)
        $ProductInstances = $Project.GetProductInstances()
        $productinstanceslist = New-Object System.Collections.ArrayList
        foreach ($Pi in $ProductInstances)
        {
            $targetslist = New-Object System.Collections.ArrayList
            foreach ($Target in $Pi.GetTargets())
            {
                $targetslist.Add((New-ProductInstanceTarget $Target.Name $Target.Key $Target.Machine.Name)) | Out-Null
            }
            $productinstanceslist.Add((New-ProductInstance $Pi.Name $Pi.OSPlatform.Name $Pi.MachinePool.Name $targetslist)) | Out-Null
        }
        $projectslist.Add((New-Project $Project.Name $Project.CreationTime.ToString() $Project.ModifiedTime.ToString() $Project.Info.Status.ToString() $productinstanceslist)) | Out-Null
    }
    ConvertTo-Json @($projectslist) -Depth $MaxJsonDepth -Compress
}

# 创建测试项目。
function Create-Project
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$project, [Parameter(Position = 2)][String]$isWindowsDriverProject)
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "please provide a project's name"
    }
    if ( $Manager.GetProjectNames().Contains($project))
    {
        throw "a project with the name $project already exists"
    }
    if ($isWindowsDriverProject -eq "true")
    {
        ConvertTo-JSON -Depth $MaxJsonDepth $Manager.CreateProject($project, $true)
    }
    else
    {
        ConvertTo-JSON -Depth $MaxJsonDepth $Manager.CreateProject($project)
    }
}

# 删除测试项目。
function Delete-Project
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$project)
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    ConvertTo-JSON -Depth $MaxJsonDepth  $Manager.DeleteProject($project)
}

# 项目加载测试目标。
function Create-Project-Target
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$target, [Parameter(Position = 2)][String]$project, [Parameter(Position = 3)][String]$machine, [Parameter(Position = 4)][String]$pool)
    if ( [String]::IsNullOrEmpty($target))
    {
        throw "Please provide a target's key."
    }
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Did not find pool $pool in Root pool, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "The test machine was not found, aborting..."
    }
    if (-Not ($WntdTarget = $WntdMachine.GetTestTargets() | Where-Object { $_.Key -eq $target }))
    {
        throw "A target that matches the target's key given was not found in the specified machine, aborting..."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    $WntdProject = $Manager.GetProject($project)
    $CreatedPI = $false
    if (-Not ($WntdPI = $WntdProject.GetProductInstances() | Where-Object { $_.OSPlatform -eq $WntdMachine.OSPlatform }))
    {
        if (-Not $WntdProject.CanCreateProductInstance($WntdMachine.OSPlatform.Description, $WntdPool, $WntdMachine.OSPlatform))
        {
            throw "can't create the project's product instance, it may be due to the project having another product instance that matches the wanted machine's pool or platform"
        }
        $WntdPI = $WntdProject.CreateProductInstance($WntdMachine.OSPlatform.Description, $WntdPool, $WntdMachine.OSPlatform)
        $CreatedPI = $true
    }
    try
    {
        $WntdPITargets = $WntdPI.GetTargets()
        if (($WntdTarget.TargetType -eq "System") -and ($WntdPITargets | Where-Object { $_.TargetType -ne "System" }))
        {
            throw "The project already has non-system targets, can't mix system and non-system targets, aborting..."
        }
        if (($WntdTarget.TargetType -ne "System") -and ($WntdPITargets | Where-Object { $_.TargetType -eq "System" }))
        {
            throw "The project already has system targets, can't mix system and non-system targets, aborting..."
        }
        $WntdtoTarget = New-Object System.Collections.ArrayList
        if ($WntdTarget.TargetType -eq "TargetCollection")
        {
            foreach ($toTarget in $WntdPI.FindTargetFromContainer($WntdTarget.ContainerId))
            {
                if ( $toTarget.Machine.Equals($WntdMachine))
                {
                    $WntdtoTarget.Add($toTarget) | Out-Null
                }
            }
        }
        else
        {
            $WntdtoTarget.Add($WntdTarget) | Out-Null
        }
        if ($WntdtoTarget.Count -lt 1)
        {
            throw "No targets to create were found, aborting..."
        }
        foreach ($toTarget in $WntdtoTarget)
        {
            if ($WntdPITargets | Where-Object { ($_.Key -eq $toTarget.Key) -and $_.Machine.Equals($toTarget.Machine) })
            {
                continue
            }
            switch ($toTarget.TargetType)
            {
                "Filter" {
                    [String[]]$HardwareIds = $toTarget.Key
                }
                "System" {
                    [String[]]$HardwareIds = "[SYSTEM]"
                }
                default {
                    [String[]]$HardwareIds = $toTarget.HardwareId
                }
            }
            if (-Not ($WntdDeviceFamily = $Manager.GetDeviceFamilies() | Where-Object { $_.Name -eq $HardwareIds[0] }))
            {
                $WntdDeviceFamily = $Manager.CreateDeviceFamily($HardwareIds[0], $HardwareIds)
            }
            if ($WntdPITargets | Where-Object { ($_.Key -eq $toTarget.Key) })
            {
                $WntdTargetFamily = ($WntdPITargets | Where-Object { ($_.Key -eq $toTarget.Key) })[0].TargetFamily
            }
            else
            {
                $WntdTargetFamily = $WntdPI.CreateTargetFamily($WntdDeviceFamily)
            }
            $WntdTargetFamily.CreateTarget($toTarget) | Out-Null
        }
    }
    catch
    {
        if ($CreatedPI)
        {
            $WntdProject.DeleteProductInstance($WntdMachine.OSPlatform.Description)
        }
        throw
    }

    return "{}"
}

# 项目移除测试目标。
function Delete-Project-Target
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$target, [Parameter(Position = 2)][String]$project, [Parameter(Position = 3)][String]$machine, [Parameter(Position = 4)][String]$pool)
    if ( [String]::IsNullOrEmpty($target))
    {
        throw "Please provide a target's key."
    }
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Did not find pool $pool in Root pool, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "The test machine was not found, aborting..."
    }
    if (-Not ($WntdTarget = $WntdMachine.GetTestTargets() | Where-Object { $_.Key -eq $target }))
    {
        throw "A target that matches the target's key given was not found in the specified machine, aborting..."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    else
    {
        $WntdProject = $Manager.GetProject($project)
    }
    if (-Not ($WntdPI = $WntdProject.GetProductInstances() | Where-Object { $_.OSPlatform -eq $WntdMachine.OSPlatform }))
    {
        throw "Machine pool not targeted in the project."
    }
    $WntdtoDelete = New-Object System.Collections.ArrayList
    if ($WntdTarget.TargetType -eq "TargetCollection")
    {
        foreach ($toDelete in $WntdPI.FindTargetFromContainer($WntdTarget.ContainerId))
        {
            $WntdPI.GetTargets() | Where-Object {
                ($_.Key -eq $toDelete.Key) -and ($_.Machine.Equals($toDelete.Machine))
            } | ForEach-Object {
                $WntdtoDelete.Add($_) | Out-Null
            }
        }
    }
    else
    {
        $WntdtoDelete.Add($WntdTarget) | Out-Null
    }
    foreach ($toDelete in $WntdtoDelete)
    {
        $WntdPI.DeleteTarget($toDelete.Key, $toDelete.Machine)
    }
    if ($WntdPI.GetTargets().Count -lt 1)
    {
        $WntdProject.DeleteProductInstance($WntdPI.Name)
    }

    return "{}"
}

function Parse-Schedule-Options
{
    [CmdletBinding()]
    param([Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption] $scheduleoptions)
    $ParsedScheduleOptions = New-Object System.Collections.ArrayList
    if (($scheduleoptions -band [Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::RequiresMultipleMachines) -eq [Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::RequiresMultipleMachines)
    {
        $ParsedScheduleOptions.Add([Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::RequiresMultipleMachines.ToString()) | Out-Null
    }
    if (($scheduleoptions -band [Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ScheduleOnAllTargets) -eq [Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ScheduleOnAllTargets)
    {
        $ParsedScheduleOptions.Add([Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ScheduleOnAllTargets.ToString()) | Out-Null
    }
    if (($scheduleoptions -band [Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ScheduleOnAnyTarget) -eq [Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ScheduleOnAnyTarget)
    {
        $ParsedScheduleOptions.Add([Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ScheduleOnAnyTarget.ToString()) | Out-Null
    }
    if (($scheduleoptions -band [Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ConsolidateScheduleAcrossTargets) -eq [Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ConsolidateScheduleAcrossTargets)
    {
        $ParsedScheduleOptions.Add([Microsoft.Windows.Kits.Hardware.ObjectModel.DistributionOption]::ConsolidateScheduleAcrossTargets.ToString()) | Out-Null
    }
    return ,$ParsedScheduleOptions
}

# 列出所有测试项。
function List-Tests
{
    [CmdletBinding()]
    param([Switch]$manual, [Switch]$auto, [Switch]$failed, [Switch]$inqueue, [Switch]$notrun, [Switch]$passed, [Switch]$running, [String]$playlist, [Parameter(Position = 1)][String]$target, [Parameter(Position = 2)][String]$project, [Parameter(Position = 3)][String]$machine, [Parameter(Position = 4)][String]$pool)
    if ( [String]::IsNullOrEmpty($target))
    {
        throw "Please provide a target's key."
    }
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($manual -or $auto))
    {
        $manual = $true
        $auto = $true
    }
    if (-Not ($notrun -or $failed -or $passed -or $running -or $inqueue))
    {
        $notrun = $true
        $failed = $true
        $passed = $true
        $running = $true
        $inqueue = $true
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Did not find pool $pool in Root pool, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "The test machine was not found, aborting..."
    }
    if (-Not ($WntdTarget = $WntdMachine.GetTestTargets() | Where-Object { $_.Key -eq $target }))
    {
        throw "A target that matches the target's key given was not found in the specified machine, aborting..."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    $WntdProject = $Manager.GetProject($project)
    $WntdPI = $WntdProject.GetProductInstances()[0]
    <#if (-Not ($WntdPI = $WntdProject.GetProductInstances() | Where-Object { $_.OSPlatform -eq $WntdMachine.OSPlatform }))
    {
        throw "Machine pool not targeted in the project."
    }#>
    $WntdPITargets = New-Object System.Collections.ArrayList
    if ($WntdTarget.TargetType -eq "TargetCollection")
    {
        foreach ($tTarget in $WntdPI.FindTargetFromContainer($WntdTarget.ContainerId))
        {
            $WntdPI.GetTargets() | Where-Object { ($_.Key -eq $tTarget.Key) -and ($_.Machine.Equals($tTarget.Machine)) } | ForEach-Object {
                $WntdPITargets.Add($_) | Out-Null
            }
        }
        if ($WntdPITargets.Count -lt 1)
        {
            throw "The target is not being targeted by the project."
        }
    }
    else
    {
        if (-Not ($WntdPITarget = $WntdPI.GetTargets() | Where-Object { ($_.Key -eq $WntdTarget.Key) -and ($_.Machine.Equals($WntdMachine)) }))
        {
            throw "The target is not being targeted by the project."
        }
        $WntdPITargets.Add($WntdPITarget) | Out-Null
    }
    $WntdTests = New-Object System.Collections.ArrayList
    if (-Not [String]::IsNullOrEmpty($playlist))
    {
        $PlaylistManager = New-Object Microsoft.Windows.Kits.Hardware.ObjectModel.PlaylistManager $WntdProject
        $WntdPlaylist = [Microsoft.Windows.Kits.Hardware.ObjectModel.PlaylistManager]::DeserializePlaylist($playlist)
        foreach ($tTest in $PlaylistManager.GetTestsFromProjectThatMatchPlaylist($WntdPlaylist))
        {
            if ($tTest.GetTestTargets() | Where-Object { $WntdPITargets.Contains($_) })
            {
                $WntdTests.Add($tTest) | Out-Null
            }
        }
    }
    else
    {
        $WntdPITargets | ForEach-Object { $WntdTests.AddRange($_.GetTests()) }
    }
    $testslist = New-Object System.Collections.ArrayList
    foreach ($tTest in $WntdTests)
    {
        if (-Not (($manual -and ($tTest.TestType -eq "Manual")) -or ($auto -and ($tTest.TestType -eq "Automated"))))
        {
            continue
        }
        elseif (-Not (($notrun -and ($tTest.Status -eq "NotRun")) -or ($failed -and ($tTest.Status -eq "Failed")) -or ($passed -and ($tTest.Status -eq "Passed")) -or ($running -and ($tTest.ExecutionState -eq "Running")) -or ($inqueue -and ($tTest.ExecutionState -eq "InQueue"))))
        {
            continue
        }
        $testslist.Add((New-Test $tTest.Name $tTest.Id $tTest.TestType.ToString() $tTest.EstimatedRuntime.ToString() $tTest.RequiresSpecialConfiguration.ToString() $tTest.RequiresSupplementalContent.ToString() (Parse-Schedule-Options($tTest.ScheduleOptions)) $tTest.Status.ToString() $tTest.ExecutionState.ToString())) | Out-Null
    }
    return (ConvertTo-Json @($testslist) -Depth $MaxJsonDepth -Compress)
}

function Load-Play-List
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$project, [Parameter(Position = 2)][String]$playlist)
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($playlist))
    {
        throw "Please provide a path to a playlist file."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    else
    {
        $WntdProject = $Manager.GetProject($project)
    }
    $PlaylistManager = New-Object Microsoft.Windows.Kits.Hardware.ObjectModel.PlaylistManager($WntdProject)
    ConvertTo-JSON -Depth $MaxJsonDepth $PlaylistManager.LoadPlaylist($playlist) | Out-Null
}

# 获取测试项信息。
function Get-Test-Info
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$test, [Parameter(Position = 2)][String]$target, [Parameter(Position = 3)][String]$project, [Parameter(Position = 4)][String]$machine, [Parameter(Position = 5)][String]$pool)
    if ( [String]::IsNullOrEmpty($test))
    {
        throw "Please provide a test's id."
    }
    if ( [String]::IsNullOrEmpty($target))
    {
        throw "Please provide a target's key."
    }
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object {
        $_.Name -eq $pool
    }))
    {
        throw "Did not find pool $pool in Root pool, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object {
        $_.Name -eq $machine
    }))
    {
        throw "The test machine was not found, aborting..."
    }
    if (-Not ($WntdTarget = $WntdMachine.GetTestTargets() | Where-Object {
        $_.Key -eq $target
    }))
    {
        throw "A target that matches the target's key given was not found in the specified machine, aborting..."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    else
    {
        $WntdProject = $Manager.GetProject($project)
    }
    $WntdPI = $WntdProject.GetProductInstances()[0]
    <#if (-Not ($WntdPI = $WntdProject.GetProductInstances() | Where-Object {
        $_.OSPlatform -eq $WntdMachine.OSPlatform
    }))
    {
        throw "Machine pool not targeted in the project."
    }#>
    $WntdPITargets = New-Object System.Collections.ArrayList
    if ($WntdTarget.TargetType -eq "TargetCollection")
    {
        foreach ($tTarget in $WntdPI.FindTargetFromContainer($WntdTarget.ContainerId))
        {
            $WntdPI.GetTargets() | Where-Object {
                ($_.Key -eq $tTarget.Key) -and ($_.Machine.Equals($tTarget.Machine))
            } | ForEach-Object {
                $WntdPITargets.Add($_) | Out-Null
            }
        }
        if ($WntdPITargets.Count -lt 1)
        {
            throw "The target is not being targeted by the project."
        }
    }
    else
    {
        if (-Not ($WntdPITarget = $WntdPI.GetTargets() | Where-Object {
            ($_.Key -eq $WntdTarget.Key) -and ($_.Machine.Equals($WntdMachine))
        }))
        {
            throw "The target is not being targeted by the project."
        }
        $WntdPITargets.Add($WntdPITarget) | Out-Null
    }
    $WntdTests = New-Object System.Collections.ArrayList
    $WntdPITargets | ForEach-Object {
        $WntdTests.AddRange($_.GetTests())
    }
    if (-Not ($WntdTest = $WntdTests | Where-Object {
        $_.Id -eq $test
    }))
    {
        throw "Didn't find a test with the id given."
    }
    @((New-Test $WntdTest.Name $WntdTest.Id $WntdTest.TestType.ToString() $WntdTest.EstimatedRuntime.ToString() $WntdTest.RequiresSpecialConfiguration.ToString() $WntdTest.RequiresSupplementalContent.ToString() (Parse-Schedule-Options($WntdTest.ScheduleOptions)) $WntdTest.Status.ToString() $WntdTest.ExecutionState.ToString())) | ConvertTo-Json -Depth $MaxJsonDepth -Compress
}

# 开启测试。
function Queue-Test
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$test, [Parameter(Position = 2)][String]$target, [Parameter(Position = 3)][String]$project, [Parameter(Position = 4)][String]$machine, [Parameter(Position = 5)][String]$pool)
    if ( [String]::IsNullOrEmpty($test))
    {
        throw "Please provide a test's id."
    }
    if ( [String]::IsNullOrEmpty($target))
    {
        throw "Please provide a target's key."
    }
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Did not find pool $pool in Root pool, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "The test machine was not found, aborting..."
    }
    if (-Not ($WntdTarget = $WntdMachine.GetTestTargets() | Where-Object { $_.Key -eq $target }))
    {
        throw "A target that matches the target's key given was not found in the specified machine, aborting..."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    $WntdProject = $Manager.GetProject($project)
    $WntdPI = $WntdProject.GetProductInstances()[0]
    <#if (-Not ($WntdPI = $WntdProject.GetProductInstances() | Where-Object { $_.OSPlatform -eq $WntdMachine.OSPlatform }))
    {
        throw "Machine pool not targeted in the project."
    }#>
    $WntdPITargets = New-Object System.Collections.ArrayList
    if ($WntdTarget.TargetType -eq "TargetCollection")
    {
        foreach ($tTarget in $WntdPI.FindTargetFromContainer($WntdTarget.ContainerId))
        {
            $WntdPI.GetTargets() | Where-Object { ($_.Key -eq $tTarget.Key) -and ($_.Machine.Equals($tTarget.Machine)) } | ForEach-Object {
                $WntdPITargets.Add($_) | Out-Null
            }
        }
        if ($WntdPITargets.Count -lt 1)
        {
            throw "The target is not being targeted by the project."
        }
    }
    else
    {
        if (-Not ($WntdPITarget = $WntdPI.GetTargets() | Where-Object { ($_.Key -eq $WntdTarget.Key) -and ($_.Machine.Equals($WntdMachine)) }))
        {
            throw "The target is not being targeted by the project."
        }
        $WntdPITargets.Add($WntdPITarget) | Out-Null
    }
    $WntdTests = New-Object System.Collections.ArrayList
    $WntdPITargets | ForEach-Object {
        $WntdTests.AddRange($_.GetTests())
    }
    if (-Not ($WntdTest = $WntdTests | Where-Object { $_.Id -eq $test }))
    {
        throw "Didn't find a test with the id given."
    }
    $WntdTest.QueueTest() | Out-Null

    return "{}"
}

function Apply-Project-Filters
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$project)
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    else
    {
        $WntdProject = $Manager.GetProject($project)
    }
    $WntdFilterEngine = New-Object Microsoft.Windows.Kits.Hardware.FilterEngine.DatabaseFilterEngine $Manager
    $WntdFilterResultDictionary = $WntdFilterEngine.Filter($WntdProject)
    $Count = 0
    foreach ($tFilterResultCollection in $WntdFilterResultDictionary.Values)
    {
        $Count += $tFilterResultCollection.Count
    }
    @(New-FilterResult $Count) | ConvertTo-Json -Depth $MaxJsonDepth -Compress
}

function Apply-Test-Result-Filters
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$result, [Parameter(Position = 2)][String]$test, [Parameter(Position = 3)][String]$target, [Parameter(Position = 4)][String]$project, [Parameter(Position = 5)][String]$machine, [Parameter(Position = 6)][String]$pool)
    if ( [String]::IsNullOrEmpty($result))
    {
        throw "Please provide a test result's index."
    }
    if ( [String]::IsNullOrEmpty($test))
    {
        throw "Please provide a test's id."
    }
    if ( [String]::IsNullOrEmpty($target))
    {
        throw "Please provide a target's key."
    }
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object {
        $_.Name -eq $pool
    }))
    {
        throw "Did not find pool $pool in Root pool, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object {
        $_.Name -eq $machine
    }))
    {
        throw "The test machine was not found, aborting..."
    }
    if (-Not ($WntdTarget = $WntdMachine.GetTestTargets() | Where-Object {
        $_.Key -eq $target
    }))
    {
        throw "A target that matches the target's key given was not found in the specified machine, aborting..."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    else
    {
        $WntdProject = $Manager.GetProject($project)
    }
    if (-Not ($WntdPI = $WntdProject.GetProductInstances() | Where-Object {
        $_.OSPlatform -eq $WntdMachine.OSPlatform
    }))
    {
        throw "Machine pool not targeted in the project."
    }
    $WntdPITargets = New-Object System.Collections.ArrayList
    if ($WntdTarget.TargetType -eq "TargetCollection")
    {
        foreach ($tTarget in $WntdPI.FindTargetFromContainer($WntdTarget.ContainerId))
        {
            $WntdPI.GetTargets() | Where-Object {
                ($_.Key -eq $tTarget.Key) -and ($_.Machine.Equals($tTarget.Machine))
            } | ForEach-Object {
                $WntdPITargets.Add($_) | Out-Null
            }
        }
        if ($WntdPITargets.Count -lt 1)
        {
            throw "The target is not being targeted by the project."
        }
    }
    else
    {
        if (-Not ($WntdPITarget = $WntdPI.GetTargets() | Where-Object {
            ($_.Key -eq $WntdTarget.Key) -and ($_.Machine.Equals($WntdMachine))
        }))
        {
            throw "The target is not being targeted by the project."
        }
        $WntdPITargets.Add($WntdPITarget) | Out-Null
    }
    $WntdTests = New-Object System.Collections.ArrayList
    $WntdPITargets | ForEach-Object {
        $WntdTests.AddRange($_.GetTests())
    }
    if (-Not ($WntdTest = $WntdTests | Where-Object {
        $_.Id -eq $test
    }))
    {
        throw "Didn't find a test with the id given."
    }
    if (-Not ($WntdTest.GetTestResults().Count -ge 1))
    {
        throw "The test hasen't been queued, can't find test results."
    }
    else
    {
        $WntdResult = $WntdTest.GetTestResults()[$result]
    }
    $WntdFilterEngine = New-Object Microsoft.Windows.Kits.Hardware.FilterEngine.DatabaseFilterEngine $Manager
    $WntdFilterResultCollection = $WntdFilterEngine.Filter($WntdResult)
    @(New-FilterResult $WntdFilterResultCollection.Count) | ConvertTo-Json -Depth $MaxJsonDepth -Compress
}

# 获取测试日志。
function List-Test-Results
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$test, [Parameter(Position = 2)][String]$target, [Parameter(Position = 3)][String]$project, [Parameter(Position = 4)][String]$machine, [Parameter(Position = 5)][String]$pool)
    if ( [String]::IsNullOrEmpty($test))
    {
        throw "Please provide a test's id."
    }
    if ( [String]::IsNullOrEmpty($target))
    {
        throw "Please provide a target's key."
    }
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Did not find pool $pool in Root pool, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "The test machine was not found, aborting..."
    }
    if (-Not ($WntdTarget = $WntdMachine.GetTestTargets() | Where-Object { $_.Key -eq $target }))
    {
        throw "A target that matches the target's key given was not found in the specified machine, aborting..."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    else
    {
        $WntdProject = $Manager.GetProject($project)
    }
    $WntdPI = $WntdProject.GetProductInstances()[0]
    <#if (-Not ($WntdPI = $WntdProject.GetProductInstances() | Where-Object { $_.OSPlatform -eq $WntdMachine.OSPlatform }))
    {
        throw "Machine pool not targeted in the project."
    }#>

    $WntdPITargets = New-Object System.Collections.ArrayList
    if ($WntdTarget.TargetType -eq "TargetCollection")
    {
        foreach ($tTarget in $WntdPI.FindTargetFromContainer($WntdTarget.ContainerId))
        {
            $WntdPI.GetTargets() | Where-Object { ($_.Key -eq $tTarget.Key) -and ($_.Machine.Equals($tTarget.Machine)) } | ForEach-Object { $WntdPITargets.Add($_) | Out-Null }
        }
        if ($WntdPITargets.Count -lt 1)
        {
            throw "The target is not being targeted by the project."
        }
    }
    else
    {
        if (-Not ($WntdPITarget = $WntdPI.GetTargets() | Where-Object { ($_.Key -eq $WntdTarget.Key) -and ($_.Machine.Equals($WntdMachine)) }))
        {
            throw "The target is not being targeted by the project."
        }
        $WntdPITargets.Add($WntdPITarget) | Out-Null
    }
    $WntdTests = New-Object System.Collections.ArrayList
    $WntdPITargets | ForEach-Object {
        $WntdTests.AddRange($_.GetTests())
    }
    if (-Not ($WntdTest = $WntdTests | Where-Object { $_.Id -eq $test }))
    {
        throw "Didn't find a test with the id given."
    }
    if (-Not ($WntdTest.GetTestResults().Count -ge 1))
    {
        throw "The test hasen't been queued, can't find test results."
    }
    $WntdResults = $WntdTest.GetTestResults()
    $testresultlist = New-Object System.Collections.ArrayList
    foreach ($tTestResult in $WntdResults)
    {
        $tTestResult.Refresh()
        $taskslist = New-Object System.Collections.ArrayList
        foreach ($tTask in $tTestResult.GetTasks())
        {
            $subtaskslist = New-Object System.Collections.ArrayList
            if ( $tTask.GetChildTasks())
            {
                foreach ($subtTask in $tTask.GetChildTasks())
                {
                    $subtasktype = (New-Task $subtTask.Name $subtTask.Stage $subtTask.Status.ToString() $subtTask.TaskErrorMessage $subtTask.TaskType (New-Object System.Collections.ArrayList))
                    $subtaskslist.Add($subtasktype) | Out-Null
                }
            }
            $tasktype = (New-Task $tTask.Name $tTask.Stage $tTask.Status.ToString() $tTask.TaskErrorMessage $tTask.TaskType $subtaskslist)
            $taskslist.Add($tasktype) | Out-Null
        }
        $testresultlist.Add((New-TestResult $tTestResult.Test.Name $tTestResult.CompletionTime.ToString() $tTestResult.ScheduleTime.ToString() $tTestResult.StartTime.ToString() $tTestResult.Status.ToString() $tTestResult.AreFiltersApplied.ToString() $tTestResult.Target.Name $taskslist)) | Out-Null
    }
    return (ConvertTo-Json @($testresultlist) -Depth $MaxJsonDepth -Compress)
}

# 将测试日志打包。
function Zip-Test-Result-Logs
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$target, [Parameter(Position = 2)][String]$project, [Parameter(Position = 3)][String]$machine, [Parameter(Position = 4)][String]$pool)
    if ( [String]::IsNullOrEmpty($target))
    {
        throw "Please provide a target's key."
    }
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "Please provide a machine's name."
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "Please provide a pool's name."
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "Did not find pool $pool in Root pool, aborting..."
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "The test machine was not found, aborting..."
    }
    if (-Not ($WntdTarget = $WntdMachine.GetTestTargets() | Where-Object { $_.Key -eq $target }))
    {
        throw "A target that matches the target's key given was not found in the specified machine, aborting..."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    $WntdProject = $Manager.GetProject($project)
    $WntdPI = $WntdProject.GetProductInstances()[0]
    <#if (-Not ($WntdPI = $WntdProject.GetProductInstances() | Where-Object { $_.OSPlatform -eq $WntdMachine.OSPlatform }))
    {
        throw "Machine pool not targeted in the project."
    }#>
    $WntdPITargets = New-Object System.Collections.ArrayList
    if ($WntdTarget.TargetType -eq "TargetCollection")
    {
        foreach ($v in $WntdPI.FindTargetFromContainer($WntdTarget.ContainerId))
        {
            $WntdPI.GetTargets() | Where-Object {
                ($_.Key -eq $v.Key) -and ($_.Machine.Equals($v.Machine))
            } | ForEach-Object {
                $WntdPITargets.Add($_) | Out-Null
            }
        }
        if ($WntdPITargets.Count -lt 1)
        {
            throw "The target is not being targeted by the project."
        }
    }
    else
    {
        if (-Not ($WntdPITarget = $WntdPI.GetTargets() | Where-Object { ($_.Key -eq $WntdTarget.Key) -and ($_.Machine.Equals($WntdMachine)) }))
        {
            throw "The target is not being targeted by the project."
        }
        $WntdPITargets.Add($WntdPITarget) | Out-Null
    }
    $WntdTests = New-Object System.Collections.ArrayList
    $WntdPITargets | ForEach-Object {
        $WntdTests.AddRange($_.GetTests())
    }

    $date = (get-date).ToString("hhmmss")
    [string]$tempPath = $env:TEMP
    $LogsDir = $tempPath + "\${project}_${date}\"
    $ZipPath = $tempPath + "\${project}_${date}.zip"
    try
    {
        for ($i = 0; $i -lt $WntdTests.Count; $i++) {
            $test = $WntdTests[$i]
            $testName = ($i + 1).ToString() + ". " + $test.Status.ToString() + " " + ($test.Name.ToString())
            $testName = (Check-File-Name $testName)
            $parentPath = [System.IO.Path]::Combine($LogsDir, $testName).ToString()
            New-Item -ItemType Directory -Force -Path $parentPath | Out-Null
            $testResults = $test.GetTestResults()
            for ($j = 0; $j -lt $testResults.Count; $j++) {
                $testResult = $testResults[$j]
                $testResult.Refresh()
                foreach ($log in $testResult.GetLogs())
                {
                    $logName = (Check-File-Name $log.Name)
                    $path = [System.IO.Path]::Combine($parentPath, $logName).ToString()
                    $log.WriteLogTo($path)
                }

                $childTasks = $testResult.GetTasks()
                Write-Tasks-Log-File $childTasks $parentPath
            }
        }
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        [IO.Compression.ZipFile]::CreateFromDirectory($LogsDir, $ZipPath)
        #Compress-Archive -Path $LogsDir -DestinationPath $ZipPath -Update
    }
    finally
    {
        Remove-Item -Recurse -Force $LogsDir
    }
    @(New-TestResultLogsZip "" "" "" $ZipPath) | ConvertTo-Json -Depth $MaxJsonDepth -Compress
}

# 创建 hLKX 包。
function Create-Project-Package
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$project, [Parameter(Position = 2)][String]$package, [Parameter(Position = 3)][String]$driverPath)
    if ( [String]::IsNullOrEmpty($project))
    {
        throw "Please provide a project's name."
    }
    if (-Not ($Manager.GetProjectNames().Contains($project)))
    {
        throw "No project with the given name was found, aborting..."
    }
    else
    {
        $WntdProject = $Manager.GetProject($project)
    }
    if (-Not [String]::IsNullOrEmpty($package))
    {
        $PackagePath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($package)
    }
    else
    {
        if (-Not (Test-Path ($env:TEMP + "\prometheus_packages\")))
        {
            New-Item ($env:TEMP + "\prometheus_packages\") -ItemType Directory | Out-Null
        }
        $PackagePath = $env:TEMP + "\prometheus_packages\" + $( get-date ).ToString("dd-MM-yyyy") + "_" + $( get-date ).ToString("hh_mm_ss") + "_" + $WntdProject.Name + ".hlkx"
    }
    $PackageWriter = New-Object Microsoft.Windows.Kits.Hardware.ObjectModel.Submission.PackageWriter $WntdProject
    $targetList = New-Object "System.Collections.Generic.List``1[Microsoft.Windows.Kits.Hardware.ObjectModel.Target]"
    $projectInfo = $Manager.GetProject($project)
    $projectInfo.GetProductInstances() | ForEach-Object {
        $targetlist.AddRange($_.GetTargets())
    }
    $localeList = New-Object "System.Collections.Generic.List``1[System.string]"
    $localeList.Add([Microsoft.Windows.Kits.Hardware.ObjectModel.ProjectManager]::GetLocaleList()[0])
    $localeList.Add([Microsoft.Windows.Kits.Hardware.ObjectModel.ProjectManager]::GetLocaleList()[1])
    $localeList.Add([Microsoft.Windows.Kits.Hardware.ObjectModel.ProjectManager]::GetLocaleList()[2])
    $errorMessages = New-Object "System.Collections.Specialized.StringCollection"
    $warningMessages = New-Object "System.Collections.Specialized.StringCollection"
    if ($packageWriter.AddDriver($driverPath, [System.Management.Automation.Language.NullString]::Value,$targetList.AsReadOnly(),$localeList.AsReadOnly(), [ref]$errorMessages, [ref]$warningMessages) -eq $false)
    {
        $err = "Add driver failed to add this driver found at : $driverPath"
        foreach ($msg in $errorMessages)
        {
            $err = "$err\r\n$msg"
        }
        throw $err
    }
    if ($warningMessages.Count -ne 0)
    {
        $err = "Add driver found warnings in the package found at : $driverPath"
        foreach ($msg in $warningMessages)
        {
            $err = "$err\r\n$msg"
        }
        throw $err
    }
    $PackageWriter.Save($PackagePath)
    $PackageWriter.Dispose()
    @(New-ProjectPackage $WntdProject.Name $PackagePath) | ConvertTo-Json -Depth $MaxJsonDepth -Compress
}

function Get-Aall-Features()
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$projectName, [Parameter(Position = 2)][String]$driver)

    $project = $Manager.GetProject($projectName)
    $inst = $project.GetProductInstances()

    $tf2 = $inst.GetTargets()
    foreach ($t2 in $tf2)
    {
        if ($t2.Name -eq $driver)
        {
            $d2 = $t2
            break
        }
    }

    $tf1 = $inst.GetTargetFamilies()
    foreach ($t1 in $tf1)
    {
        if ($t1.GroupId -eq $d2.TargetFamily.GroupId)
        {
            $d1 = $t1
            break
        }
    }

    return (ConvertTo-Json -Depth $MaxJsonDepth -Compress $d1.GetFeatures())
}

function Get-Target-Features()
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$projectName, [Parameter(Position = 2)][String]$driver)

    $project = $Manager.GetProject($projectName)
    $inst = $project.GetProductInstances()
    $tf = $inst.GetTargets()
    foreach ($t in $tf)
    {
        if ($t.Name -eq $driver)
        {
            $d = $t
            break
        }
    }

    $s = $d.GetFeatures()

    ConvertTo-Json -Depth $MaxJsonDepth -Compress $s | Write-Output
}

function Add-Target-Feature()
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$projectName, [Parameter(Position = 2)][String]$driver, [Parameter(Position = 3)][String]$feature)

    $project = $Manager.GetProject($projectName)
    $inst = $project.GetProductInstances()

    $tf2 = $inst.GetTargets()
    foreach ($t2 in $tf2)
    {
        if ($t2.Name -eq $driver)
        {
            $d2 = $t2
            break
        }
    }

    $tf1 = $inst.GetTargetFamilies()
    foreach ($t1 in $tf1)
    {
        if ($t1.GroupId -eq $d2.TargetFamily.GroupId)
        {
            $d1 = $t1
            break
        }
    }

    foreach ($t3 in $d1.GetFeatures())
    {
        if ($t3.FullName -eq $feature)
        {
            $d2.AddFeature($t3)
            break
        }
    }

    return "{}"
}

function Remove-Target-Feature()
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$projectName, [Parameter(Position = 2)][String]$driver, [Parameter(Position = 3)][String]$feature)

    $project = $Manager.GetProject($projectName)
    $inst = $project.GetProductInstances()

    $tf2 = $inst.GetTargets()
    foreach ($t2 in $tf2)
    {
        if ($t2.Name -eq $driver)
        {
            $d2 = $t2
            break
        }
    }

    ConvertTo-JSON -Depth $MaxJsonDepth $d2.RemoveFeature($feature)
}

function Add-Target-All-Feature()
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$projectName, [Parameter(Position = 2)][String]$driver)

    if ( [String]::IsNullOrEmpty($projectName))
    {
        throw "Please provide a project's name."
    }
    if ( [String]::IsNullOrEmpty($driver))
    {
        throw "Please provide a driver's name."
    }
    if (-Not ($Manager.GetProjectNames().Contains($projectName)))
    {
        throw "No project with the given name was found, aborting..."
    }

    $project = $Manager.GetProject($projectName)
    $inst = $project.GetProductInstances()
    $target = $inst.GetTargets() | Where-Object { $_.Name -eq $driver }
    $selectedFeatures = $target.GetFeatures()

    foreach ($v in $Manager.GetFeatures())
    {
        if (-Not ($selectedFeatures.Contains($v)) -and $v.TargetType -eq $target.TargetType)
        {
            $target.AddFeature($v)
        }
    }

    return "{}"
}

function Get-Target()
{
    [CmdletBinding()]
    param([Parameter(Position = 1)][String]$name, [Parameter(Position = 2)][String]$machine, [Parameter(Position = 3)][String]$pool)

    if ( [String]::IsNullOrEmpty($machine))
    {
        throw "please provide a machine's name"
    }
    if ( [String]::IsNullOrEmpty($pool))
    {
        throw "please provide a pool's name"
    }
    if (-Not ($WntdPool = $RootPool.GetChildPools() | Where-Object { $_.Name -eq $pool }))
    {
        throw "did not find pool $pool in Root pool"
    }
    if (-Not ($WntdMachine = $WntdPool.GetMachines() | Where-Object { $_.Name -eq $machine }))
    {
        throw "the test machine was not found"
    }
    $targetData = $WntdMachine.GetTestTargets() | Where-Object { $_.Name -eq $name }
    $targetData = (New-Target $targetData.Name $targetData.Key $targetData.TargetType)
    return (ConvertTo-Json $targetData -Depth $MaxJsonDepth -Compress)
}

# 脚本逻辑入口。

[System.Reflection.Assembly]::LoadFrom($env:WTTSTDIO + "microsoft.windows.kits.hardware.filterengine.dll") | Out-Null
[System.Reflection.Assembly]::LoadFrom($env:WTTSTDIO + "microsoft.windows.kits.hardware.objectmodel.dll") | Out-Null
[System.Reflection.Assembly]::LoadFrom($env:WTTSTDIO + "microsoft.windows.kits.hardware.objectmodel.dbconnection.dll") | Out-Null
[System.Reflection.Assembly]::LoadFrom($env:WTTSTDIO + "microsoft.windows.kits.hardware.objectmodel.submission.dll") | Out-Null
[System.Reflection.Assembly]::LoadFrom($env:WTTSTDIO + "microsoft.windows.kits.hardware.objectmodel.submission.package.dll") | Out-Null

$ConnectFileName = $env:WTTSTDIO + "connect.xml"
$ConnectFile = [xml](Get-Content $ConnectFileName)
$ControllerName = $ConnectFile.Connection.GetAttribute("Server")
$DatabaseName = $connectFile.Connection.GetAttribute("Source")
$Manager = New-Object Microsoft.Windows.Kits.Hardware.ObjectModel.DBConnection.DatabaseProjectManager -Args $ControllerName, $DatabaseName
if ($null -eq $Manager)
{
    Fail-Host("Connecting to $ControllerName failed")
}
$MaxJsonDepth = 20
$RootPool = $Manager.GetRootMachinePool()
$AllowedCmds = New-Object System.Collections.ArrayList
$AllowedCmds.AddRange(("List-Pools", "Create-Pool", "Delete-Pool", "Move-Machine", "Set-Machine-State", "Delete-Machine",
"List-Machine-Targets", "List-Projects", "Create-Project", "Delete-Project", "Create-Project-Target", "Delete-Project-Target",
"List-Tests", "Get-Test-Info", "Queue-Test", "Apply-Project-Filters", "Apply-Test-Result-Filters", "List-Test-Results",
"Zip-Test-Result-Logs", "Create-Project-Package", "Load-Play-List", "Get-All-Features", "Get-Target-Features", "Add-Target-Feature",
"Remove-Target-Feature", "Add-Target-All-Feature", "Get-Target"))

[System.Collections.ArrayList]$cmdlinelist = $cmdline.Split(",")
$cmd = $cmdlinelist[0]
$cmdlinelist.RemoveAt(0)
$cmdargs = $cmdlinelist -join '" "'
$cmdargs = '"' + $cmdargs + '"'

if ( $AllowedCmds.Contains($cmd))
{
    try
    {
        $actionoutput = Invoke-Expression "$cmd $cmdargs"
        $output = @(New-ActionResult $actionoutput) | ConvertTo-Json -Depth $MaxJsonDepth -Compress
        $JoinedOutput = $output -join [Environment]::NewLine
        Write-Host $JoinedOutput
    }
    catch
    {
        $output = New-ActionResult $nil $_.Exception | ConvertTo-Json -Depth $MaxJsonDepth -Compress
        $JoinedOutput = $output -join [Environment]::NewLine
        Write-Host $JoinedOutput
    }
}
else
{
    Fail-Host("No such action")
}
